package wiki

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	category "gitlab.com/sazl/succession/pkg/succession/category/model"
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

	resp, ok := response.(fetchCategoryByNameResponse)
	if !ok {
		return category.Category{}
	}

	var categoryMembers []category.Category
	for _, cm := range resp.Query.CategoryMembers {
		c := category.Category{
			ID: category.ID(strconv.Itoa(cm.PageID)),
			Namespace: cm.Namespace,
			Title: category.Title(cm.Title),
			Name: category.Name(cm.Title),
		}
		categoryMembers = append(categoryMembers, c)
	}

	if len(categoryMembers) == 0 {
		return category.Category{}
	}
	firstCategory := categoryMembers[0]

	result := category.Category{
		ID: firstCategory.ID,
		Namespace: firstCategory.Namespace,
		Title: firstCategory.Title,
		Name: firstCategory.Name,
		CategoryMembers: categoryMembers,
	}


	return result
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
	Query categoryMembersResponse `json:"query"`
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
