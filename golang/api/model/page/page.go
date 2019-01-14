package page

// ID - Wiki API page ID
type ID string

// Category - Wiki API category
type Category struct {
	ID        ID
	Namespace int
	Title     string
}