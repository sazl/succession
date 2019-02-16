package main

import (
	"context"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	log "github.com/go-kit/kit/log"
	run "github.com/oklog/run"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	inmem "gitlab.com/sazl/succession/pkg/succession/category/persistence/inmem"
	categorysvc "gitlab.com/sazl/succession/pkg/succession/category/service"
	wiki "gitlab.com/sazl/succession/pkg/succession/wiki"
)

const (
	defaultServiceName    = "category_service"
	defaultHTTPPort       = "8001"
	defaultDebugPort      = "8080"
	defaultZipkinV1URL    = ""
	defaultZipkinV2URL    = ""
	defaultWikiServiceURL = "https://en.wikipedia.org/w/api.php"
)

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

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func main() {
	var (
		ctx = context.Background()
	)

	fs := flag.NewFlagSet("addsvc", flag.ExitOnError)
	var (
		serviceName = fs.String("name", defaultServiceName, "Service name")
		// debugAddr      = fs.String("debug.addr", defaultDebugPort, "Debug and metrics listen address")
		httpAddr       = fs.String("http-addr", ":"+defaultHTTPPort, "HTTP listen address")
		zipkinV2URL    = fs.String("zipkin-url", defaultZipkinV2URL, "Enable Zipkin v2 tracing (zipkin-go) using a Reporter URL e.g. http://localhost:9411/api/v2/spans")
		zipkinV1URL    = fs.String("zipkin-v1-url", defaultZipkinV1URL, "Enable Zipkin v1 tracing (zipkin-go-opentracing) using a collector URL e.g. http://localhost:9411/api/v1/spans")
		wikiServiceURL = fs.String("wiki-service-url", defaultWikiServiceURL, "Wiki service url")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
	fs.Parse(os.Args[1:])

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		stdlog.SetFlags(0)                             // flags are handled by Go kit's logger
		stdlog.SetOutput(log.NewStdlibAdapter(logger)) // redirect anything using stdlib log to us
	}

	httpLogger := log.With(logger, "component", "http")
	transportLogger := log.With(logger, "component", *serviceName)
	// traceLogger := log.With(transportLogger, "echo", "tracing")

	var otTracer stdopentracing.Tracer
	{
		if *zipkinV1URL != "" && *zipkinV2URL == "" {
			logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", *zipkinV1URL)
			collector, err := zipkinot.NewHTTPCollector(*zipkinV1URL)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}

			defer collector.Close()

			var (
				debug       = false
				hostPort    = "localhost:80"
				serviceName = *serviceName
			)

			recorder := zipkinot.NewRecorder(collector, debug, hostPort, serviceName)
			otTracer, err = zipkinot.NewTracer(recorder)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
		} else {
			logger.Log("tracer", "no-op", "type", "OpenTracing", "URL", "no-op")
			otTracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = "localhost:80"
			useNoopTracer = (*zipkinV2URL == "")
			reporter      = zipkinhttp.NewReporter(*zipkinV2URL)
		)

		defer reporter.Close()

		zEP, _ := zipkin.NewEndpoint(*serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			reporter,
			zipkin.WithLocalEndpoint(zEP),
			zipkin.WithNoopTracer(useNoopTracer),
		)

		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}

		if !useNoopTracer {
			logger.Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinV2URL)
		}
	}

	var (
		categories = inmem.NewCategoryRepository()
		fieldKeys  = []string{"method"}
	)

	var ws wiki.Service
	ws = wiki.NewProxyingMiddleware(ctx, *wikiServiceURL, otTracer, zipkinTracer)(ws)

	var cs categorysvc.Service
	cs = categorysvc.NewService(categories, ws)
	cs = categorysvc.NewLoggingService(transportLogger, cs)

	var requestCount = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "api",
		Subsystem: *serviceName,
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	var requestLatency = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "api",
		Subsystem: *serviceName,
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	cs = categorysvc.NewInstrumentingService(requestCount, requestLatency, cs)

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("static"))

	categoryServiceHandler := categorysvc.MakeHandler(httpLogger, otTracer, zipkinTracer, cs)
	mux.Handle("/category/v1/", categoryServiceHandler)
	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	var g run.Group
	{
		var (
			listener, _ = net.Listen("tcp", *httpAddr)
			addr        = listener.Addr().String()
		)
		g.Add(func() error {
			logger.Log("msg", "service start", "transport", "http", "address", addr)
			return http.Serve(listener, categoryServiceHandler)
		}, func(error) {
			listener.Close()
		})
	}
	{
		// Set-up our signal handler.
		var (
			cancelInterrupt = make(chan struct{})
			signalChannel   = make(chan os.Signal, 2)
		)
		defer close(signalChannel)

		g.Add(func() error {
			signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-signalChannel:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	logger.Log("exited", g.Run())
}
