// Package api provides the use-case of booking a cargo. Used by views
// facing an administrator.
package api

import (
	"errors"

	"gitlab.com/sazl/succession/wiki"
	"gitlab.com/sazl/succession/api/model/category"
)

// ErrInvalidArgument is returned when one or more arguments are invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// Service is the interface that provides booking methods.
type Service interface {
	FindCategoryByName(name category.Name) (CategoryView, error)
}

type service struct {
	categories  category.Repository
	wikiService wiki.Service
}

func (s *service) FindCategoryByName(name category.Name) (CategoryView, error) {
	if name == "" {
		return CategoryView{}, ErrInvalidArgument
	}

	c, err := s.categories.FindByName(name)
	if err != nil {
		return CategoryView{}, err
	}

	return assemble(c), nil
}
// NewService creates a booking service with necessary dependencies.
func NewService(categories category.Repository, ws wiki.Service) Service {
	return &service{
		categories: categories,
		wikiService: ws,
	}
}

// CategoryView is a read model for category views
type CategoryView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Namespace string `json:"namespace"`
}

func assemble(c *category.Category) CategoryView {
	return CategoryView{
		ID:        string(c.ID),
		Name:      string(c.Name),
		Title:     string(c.Title),
		Namespace: string(c.Namespace),
	}
}
