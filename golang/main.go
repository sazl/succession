package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	stdlog "log"

	log "github.com/go-kit/kit/log"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go-opentracing"
	kitot "github.com/go-kit/kit/tracing/opentracing"

	categorysvc "gitlab.com/sazl/succession/api/category/service"
	inmem "gitlab.com/sazl/succession/api/category/persistence/inmem"
	"gitlab.com/sazl/succession/wiki"
)

const (
	defaultPort           = "8080"
	defaultWikiServiceURL = "https://en.wikipedia.org/w/api.php"
	defaultTracingServiceURL = ""
	defaultTracingSamplingRate = 1.0
)

func main() {
	var (
		addr  = envString("PORT", defaultPort)
		wsURL = envString("WIKI_SERVICE_URL", defaultWikiServiceURL)
		tracingURL = envString("TRACING_SERVICE_URL", defaultZipkinURL)

		serviceName = flag.String("name", "category_service", "Service Name")
		httpAddr = flag.String("http.addr", ":" + addr, "HTTP listen address")
		wikiServiceURL = flag.String("service.wiki", wsURL, "wiki service URL")
		tracingServiceURL = flag.String("service.tracing", tracingURL, "Tracing Service URL")
		tracingSamplingRate = flag.Float64("service.tracing.samplingrate", 1.0, "Tracing Sampling Rate")

		ctx = context.Background()
	)

	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		stdlog.SetFlags(0)                             // flags are handled by Go kit's logger
		stdlog.SetOutput(log.NewStdlibAdapter(logger)) // redirect anything using stdlib log to us

		httpLogger = log.With(logger, "component", "http")
		transportLogger = log.With(logger, "component", *serviceName)
		traceLogger = log.With(transportLogger, "echo", "tracing")
	}

	var (
		categories = inmem.NewCategoryRepository()
		fieldKeys = []string{"method"}
	)

	var ws wiki.Service
	ws = wiki.NewProxyingMiddleware(ctx, *wikiServiceURL)(ws)

	var cs categorysvc.Service
	cs = categorysvc.NewService(categories, ws)

	cs = categorysvc.NewTracingService(*tracingServiceURL, *tracingSamplingRate, *serviceName, cs)

	cs = categorysvc.NewLoggingService(transportLogger, cs)
	cs = categorysvc.NewLoggingService(traceLogger, cs)

	cs = categorysvc.NewInstrumentingService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: *serviceName,
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: *serviceName,
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		cs,
	)

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))

	mux.Handle("/category/v1/", categorysvc.MakeHandler(cs, httpLogger))
	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServe(*httpAddr, nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
