package categorysvc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	// kitot "github.com/go-kit/kit/tracing/opentracing"
	zipkin "github.com/go-kit/kit/tracing/zipkin"

	category "gitlab.com/sazl/succession/pkg/succession/category/model"
)

// MakeHandler returns a handler for the booking service.
func MakeHandler(logger kitlog.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, bs Service) http.Handler {
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer)

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}

	findCategoryByNameHandler := kithttp.NewServer(
		makeFindCategoryByNameEndpoint(bs),
		decodeFindCategoryByNameRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()
	r.Handle("/category/v1/{name}", findCategoryByNameHandler).Methods("GET")
	return r
}

var errBadRoute = errors.New("bad route")

func decodeFindCategoryByNameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, errBadRoute
	}
	return findCategoryByNameRequest{Name: category.Name(name)}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case category.ErrUnknown:
		w.WriteHeader(http.StatusNotFound)
	case ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
