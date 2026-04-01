package product

import "time"

type Product struct {
	AppName            string    `json:"app_name"`
	Org                string    `json:"org"`
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Category           string    `json:"category"`
	Image              string    `json:"image"`
	Price              float64   `json:"price"`
	OriginalPrice      float64   `json:"original_price"`
	DiscountPercentage float64   `json:"discount_percentage"`
	StockQuantity      int       `json:"stock_quantity"`
	IsNew              bool      `json:"is_new"`
	IsBestSeller       bool      `json:"is_best_seller"`
	IsOnSale           bool      `json:"is_on_sale"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	CreatedBy          string    `json:"created_by"`
	UpdatedBy          string    `json:"updated_by"`
	Status             string    `json:"status"`
}

func NewProduct(id int, org, appName string) *Product {
	return &Product{
		ID:      id,
		Org:     org,
		AppName: appName,
		Status:  "active",
	}
}
