package product

type Repository interface {
	FindAll(org string) ([]Product, error)
	FindByID(org string, id int) (*Product, error)
	FindByCategory(org, category string) ([]Product, error)
	Save(p *Product) error
	Update(p *Product) error
	Delete(org string, id int) error
	NextID(org string) (int, error)
	Exists(org string, id int) (bool, error)
}
