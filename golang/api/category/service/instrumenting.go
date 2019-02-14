package categorysvc

import (
	"time"

	"github.com/go-kit/kit/metrics"

	category "gitlab.com/sazl/succession/api/category/model"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(counter metrics.Counter, latency metrics.Histogram, s Service) Service {
	return &instrumentingService{
		requestCount:   counter,
		requestLatency: latency,
		Service:        s,
	}
}

func (s *instrumentingService) FindCategoryByName(name category.Name) (c CategoryView, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "findCategoryMembersByName").Add(1)
		s.requestLatency.With("method", "find").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.Service.FindCategoryByName(name)
}
