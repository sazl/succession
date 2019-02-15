package categorysvc

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	category "gitlab.com/sazl/succession/pkg/succession/category/model"
)

type findCategoryByNameRequest struct {
	Name category.Name
}

type findCategoryByNameResponse struct {
	Category *CategoryView `json:"category,omitempty"`
	Err      error         `json:"error,omitempty"`
}

func (r findCategoryByNameResponse) error() error { return r.Err }

func makeFindCategoryByNameEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(findCategoryByNameRequest)
		c, err := s.FindCategoryByName(req.Name)
		return findCategoryByNameResponse{Category: &c, Err: err}, nil
	}
}
