package category

// Repository category repository
type Repository interface {
	Store(cargo *Category) error
	FindById(id ID) (*Category, error)
	FindByName(name Name) (*Category, error)
	FindAll() []*Category
}