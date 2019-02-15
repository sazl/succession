package categorysvc

import (
	"time"

	"github.com/go-kit/kit/log"

	category "gitlab.com/sazl/succession/pkg/succession/category/model"
)

type loggingService struct {
	logger log.Logger
	Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) FindCategoryByName(name category.Name) (c CategoryView, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "find",
			"name", name,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.FindCategoryByName(name)
}
