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


// Category - Wiki API category
type Category struct {
	ID		  ID
	Namespace int
	PageID	  int
	Title     Title
	Name      Name
}

// ErrUnknown unknown error
var ErrUnknown = errors.New("unknown cargo")
