package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"gitlab.com/sazl/succession/api/model/category"
)

type findCategoryByNameRequest struct {
	Name category.Name
}

type findCategoryByNameResponse struct {
	Category *CategoryView `json:"category,omitempty"`
	Err       error        `json:"error,omitempty"`
}

func (r findCategoryByNameResponse) error() error { return r.Err }

func makeFindCategoryByNameEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(findCategoryByNameRequest)
		c, err := s.FindCategoryByName(req.Name)
		return findCategoryByNameResponse{Category: &c, Err: err}, nil
	}
}
