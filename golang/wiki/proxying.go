package wiki

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"gitlab.com/sazl/succession/api/model/category"
)

type proxyService struct {
	context.Context
	FetchCategoryByNameEndpoint endpoint.Endpoint
	Service
}

func (s proxyService) FetchCategoryByName(name category.Name) category.Category {
	response, err := s.FetchCategoryByNameEndpoint(s.Context, fetchCategoryByNameRequest{
		Name: string(name),
	})
	if err != nil {
		return category.Category{}
	}

	resp := response.(fetchCategoryByNameResponse)

	return &category.Category{
		ID: 1,
		Name: "string",
		Title: "hello",
	}
}

// ServiceMiddleware defines a middleware for a routing service.
type ServiceMiddleware func(Service) Service

// NewProxyingMiddleware returns a new instance of a proxying middleware.
func NewProxyingMiddleware(ctx context.Context, proxyURL string) ServiceMiddleware {
	return func(next Service) Service {
		var e endpoint.Endpoint
		e = makeFetchCategoryByNameEndpoint(ctx, proxyURL)
		e = circuitbreaker.Hystrix("fetch-category-by-name")(e)
		return proxyService{ctx, e, next}
	}
}

type fetchCategoryByNameRequest struct {
	Name string
}

type categoryMembers struct {
	PageID    int    `json:"pageid"`
	Title     string `json:"title"`
	Namespace int    `json:"ns"`
}

type categoryMembersResponse struct {
	CategoryMembers []categoryMembers
}

type fetchCategoryByNameResponse struct {
	BatchComplete string `json:"batchcomplete"`
	Query categoryMembersResponse `json:"categorymembers"`
}

func makeFetchCategoryByNameEndpoint(ctx context.Context, instance string) endpoint.Endpoint {
	u, err := url.Parse(instance)
	if err != nil {
		panic(err)
	}
	if u.Path == "" {
		u.Path = "/paths"
	}
	return kithttp.NewClient(
		"GET", u,
		encodeFetchCategoryByNameRequest,
		decodeFetchCategoryByNameResponse,
	).Endpoint()
}

func decodeFetchCategoryByNameResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response fetchCategoryByNameResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func encodeFetchCategoryByNameRequest(_ context.Context, r *http.Request, request interface{}) error {
	req := request.(fetchCategoryByNameRequest)

	vals := r.URL.Query()
	vals.Add("action", "query")
	vals.Add("list", "categorymembers")
	vals.Add("cmtitle", "Category:" + req.Name)
	vals.Add("format", "json")
	vals.Add("cmlimit", "100")
	r.URL.RawQuery = vals.Encode()

	return nil
}
