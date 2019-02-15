// Package categorysvc provides the use-case of booking a cargo. Used by views
// facing an administrator.
package categorysvc

import (
	"fmt"
	"errors"

	category "gitlab.com/sazl/succession/pkg/succession/category/model"
	"gitlab.com/sazl/succession/pkg/succession/wiki"
)

// ErrInvalidArgument is returned when one or more arguments are invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// Service is the interface that provides booking methods.
type Service interface {
	FindCategoryByName(name category.Name) (CategoryView, error)
}

type service struct {
	categories  Repository
	wikiService wiki.Service
}

func (s *service) FindCategoryByName(name category.Name) (CategoryView, error) {
	if name == "" {
		return CategoryView{}, ErrInvalidArgument
	}

	c, err := s.categories.FindByName(name)

	if err != nil {
		cat := s.wikiService.FetchCategoryByName(name)
		cat.Name = name
		c, err = s.categories.Store(&cat)
		fmt.Println("From Remote")
		return assemble(c), err
	}

	fmt.Println("From Cache")
	return assemble(c), nil
}

// NewService creates a booking service with necessary dependencies.
func NewService(categories Repository, ws wiki.Service) Service {
	return &service{
		categories: categories,
		wikiService: ws,
	}
}

// CategoryView is a read model for category views
type CategoryView struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Title           string         `json:"title"`
	Namespace       int            `json:"namespace"`
	CategoryMembers []CategoryView `json:"categoryMembers"`
}

func assemble(c *category.Category) CategoryView {
	categoryMembers := make([]CategoryView, len(c.CategoryMembers))
	for i, c := range c.CategoryMembers {
		categoryMembers[i] = assemble(&c)
	}

	return CategoryView{
		ID:              string(c.ID),
		Name:            string(c.Name),
		Title:           string(c.Title),
		Namespace:       c.Namespace,
		CategoryMembers: categoryMembers,
	}
}
