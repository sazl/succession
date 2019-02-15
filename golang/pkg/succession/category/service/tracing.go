package categorysvc

import (
	"time"
	"net"
	"fmt"
	"os"
	stdlog "log"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go-opentracing"
	// kitot "github.com/go-kit/kit/tracing/opentracing"

	// category "gitlab.com/sazl/succession/api/category/model"
)

type tracingService struct {
	tracer opentracing.Tracer
	Service
}

func getLocalIP() (ipaddr string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("Unable to get local IP")

}


// NewTracingService returns an instance of a tracing Service
func NewTracingService(tracingServiceURL string, tracingSamplingRate float64, serviceName string, s Service) Service {
	Error := stdlog.New(os.Stdout, "ERROR:", stdlog.Ldate|stdlog.Ltime|stdlog.Lshortfile)
	var tracer opentracing.Tracer
	{
		switch {
		case tracingServiceURL != "":
			timeout := time.Second
			collector, err := zipkintracer.NewScribeCollector(tracingServiceURL, timeout, zipkintracer.ScribeBatchSize(1))

			if err != nil {
				stdlog.Fatal(err)
			}
			ipaddr, _ := getLocalIP()
			recorder := zipkintracer.NewRecorder(collector, true, ipaddr, serviceName)
			sampler := zipkintracer.NewCountingSampler(tracingSamplingRate)
			tracer, err = zipkintracer.NewTracer(recorder, zipkintracer.WithSampler(sampler), zipkintracer.WithLogger(zipkintracer.LogWrapper(Error)))
			if err != nil {
				stdlog.Fatal(err)
			}
		default:
			tracer = opentracing.GlobalTracer()
		}
	}
	return &tracingService{
		tracer: tracer,
		Service: s,
	}
}