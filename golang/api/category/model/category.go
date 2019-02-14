package category

import (
	"errors"
)

// ID - Wiki API category ID
type ID string

// Name - Wiki API category Name
type Name string

// Title - Wiki API category Name
type Title string

// Category - Wiki API Category
type Category struct {
	ID		        ID
	Namespace       int
	Title           Title
	Name            Name
	CategoryMembers []Category
}

// ErrUnknown unknown error
var ErrUnknown = errors.New("unknown cargo")
