package categorysvc

import (
	category "gitlab.com/sazl/succession/pkg/succession/category/model"
)

// Repository category repository
type Repository interface {
	FindByName(name category.Name) (*category.Category, error)
	Store(c *category.Category) (*category.Category, error)
}
