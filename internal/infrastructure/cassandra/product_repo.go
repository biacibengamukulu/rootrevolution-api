package cassandra

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"

	"rootrevolution-api/internal/domain/product"
)

type productRepository struct {
	session *gocql.Session
}

func NewProductRepository(session *gocql.Session) product.Repository {
	return &productRepository{session: session}
}

func (r *productRepository) FindAll(org string) ([]product.Product, error) {
	iter := r.session.Query(`SELECT org, id, app_name, name, description, category, image,
		price, original_price, discount_pct, stock_qty,
		is_new, is_best_seller, is_on_sale,
		created_at, updated_at, created_by, updated_by, status
		FROM products WHERE org = ? ALLOW FILTERING`, org).Iter()

	var products []product.Product
	var p product.Product
	var updatedAt *time.Time

	for iter.Scan(
		&p.Org, &p.ID, &p.AppName, &p.Name, &p.Description, &p.Category, &p.Image,
		&p.Price, &p.OriginalPrice, &p.DiscountPercentage, &p.StockQuantity,
		&p.IsNew, &p.IsBestSeller, &p.IsOnSale,
		&p.CreatedAt, &updatedAt, &p.CreatedBy, &p.UpdatedBy, &p.Status,
	) {
		if updatedAt != nil {
			p.UpdatedAt = *updatedAt
		}
		products = append(products, p)
		p = product.Product{}
		updatedAt = nil
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("querying products: %w", err)
	}

	return products, nil
}

func (r *productRepository) FindByID(org string, id int) (*product.Product, error) {
	var p product.Product
	var updatedAt *time.Time

	err := r.session.Query(`SELECT org, id, app_name, name, description, category, image,
		price, original_price, discount_pct, stock_qty,
		is_new, is_best_seller, is_on_sale,
		created_at, updated_at, created_by, updated_by, status
		FROM products WHERE org = ? AND id = ?`, org, id).Scan(
		&p.Org, &p.ID, &p.AppName, &p.Name, &p.Description, &p.Category, &p.Image,
		&p.Price, &p.OriginalPrice, &p.DiscountPercentage, &p.StockQuantity,
		&p.IsNew, &p.IsBestSeller, &p.IsOnSale,
		&p.CreatedAt, &updatedAt, &p.CreatedBy, &p.UpdatedBy, &p.Status,
	)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding product %d: %w", id, err)
	}

	if updatedAt != nil {
		p.UpdatedAt = *updatedAt
	}

	return &p, nil
}

func (r *productRepository) FindByCategory(org, category string) ([]product.Product, error) {
	iter := r.session.Query(`SELECT org, id, app_name, name, description, category, image,
		price, original_price, discount_pct, stock_qty,
		is_new, is_best_seller, is_on_sale,
		created_at, updated_at, created_by, updated_by, status
		FROM products WHERE org = ? AND category = ? ALLOW FILTERING`, org, category).Iter()

	var products []product.Product
	var p product.Product
	var updatedAt *time.Time

	for iter.Scan(
		&p.Org, &p.ID, &p.AppName, &p.Name, &p.Description, &p.Category, &p.Image,
		&p.Price, &p.OriginalPrice, &p.DiscountPercentage, &p.StockQuantity,
		&p.IsNew, &p.IsBestSeller, &p.IsOnSale,
		&p.CreatedAt, &updatedAt, &p.CreatedBy, &p.UpdatedBy, &p.Status,
	) {
		if updatedAt != nil {
			p.UpdatedAt = *updatedAt
		}
		products = append(products, p)
		p = product.Product{}
		updatedAt = nil
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("querying products by category: %w", err)
	}

	return products, nil
}

func (r *productRepository) Save(p *product.Product) error {
	return r.session.Query(`INSERT INTO products
		(org, id, app_name, name, description, category, image,
		 price, original_price, discount_pct, stock_qty,
		 is_new, is_best_seller, is_on_sale,
		 created_at, updated_at, created_by, updated_by, status)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.Org, p.ID, p.AppName, p.Name, p.Description, p.Category, p.Image,
		p.Price, p.OriginalPrice, p.DiscountPercentage, p.StockQuantity,
		p.IsNew, p.IsBestSeller, p.IsOnSale,
		p.CreatedAt, p.UpdatedAt, p.CreatedBy, p.UpdatedBy, p.Status,
	).Exec()
}

func (r *productRepository) Update(p *product.Product) error {
	p.UpdatedAt = time.Now()
	return r.session.Query(`UPDATE products SET
		app_name = ?, name = ?, description = ?, category = ?, image = ?,
		price = ?, original_price = ?, discount_pct = ?, stock_qty = ?,
		is_new = ?, is_best_seller = ?, is_on_sale = ?,
		updated_at = ?, updated_by = ?, status = ?
		WHERE org = ? AND id = ?`,
		p.AppName, p.Name, p.Description, p.Category, p.Image,
		p.Price, p.OriginalPrice, p.DiscountPercentage, p.StockQuantity,
		p.IsNew, p.IsBestSeller, p.IsOnSale,
		p.UpdatedAt, p.UpdatedBy, p.Status,
		p.Org, p.ID,
	).Exec()
}

func (r *productRepository) Delete(org string, id int) error {
	return r.session.Query(`DELETE FROM products WHERE org = ? AND id = ?`, org, id).Exec()
}

func (r *productRepository) NextID(org string) (int, error) {
	iter := r.session.Query(`SELECT id FROM products WHERE org = ? ALLOW FILTERING`, org).Iter()

	maxID := 20100
	var id int
	for iter.Scan(&id) {
		if id > maxID {
			maxID = id
		}
	}

	if err := iter.Close(); err != nil {
		return 0, err
	}

	return maxID + 1, nil
}

func (r *productRepository) Exists(org string, id int) (bool, error) {
	var count int
	err := r.session.Query(`SELECT COUNT(*) FROM products WHERE org = ? AND id = ?`, org, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
