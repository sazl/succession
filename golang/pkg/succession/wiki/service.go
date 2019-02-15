// Package wiki provides the routing domain service. It does not actually
// implement the routing service but merely acts as a proxy for a separate
// bounded context.
package wiki

import (
	category "gitlab.com/sazl/succession/pkg/succession/category/model"
)

// Service provides access to an external routing service.
type Service interface {
	FetchCategoryByName(name category.Name) category.Category
}
