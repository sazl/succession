// Package inmem provides in-memory implementations of all the domain repositories.
package inmem

import (
	"sync"

	category "gitlab.com/sazl/succession/api/category/model"
	categorysvc "gitlab.com/sazl/succession/api/category/service"
)

type categoryRepository struct {
	mtx        sync.RWMutex
	categories map[category.ID]*category.Category
}

func (r *categoryRepository) FindByName(name category.Name) (*category.Category, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	for _, val := range r.categories {
		if val.Name == name {
			return val, nil
		}
	}

	return nil, category.ErrUnknown
}

func (r *categoryRepository) Store(c *category.Category) (*category.Category, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	r.categories[c.ID] = c
	return c, nil
}

// NewCategoryRepository returns a new instance of a in-memory cargo repository.
func NewCategoryRepository() categorysvc.Repository {
	c := &categoryRepository{
		categories: make(map[category.ID]*category.Category),
	}
	return c
}
