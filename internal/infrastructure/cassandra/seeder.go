package cassandra

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gocql/gocql"

	"rootrevolution-api/config"
)

type seedProduct struct {
	AppName            string  `json:"app_name"`
	Org                string  `json:"org"`
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	Category           string  `json:"category"`
	Image              string  `json:"image"`
	Price              float64 `json:"price"`
	OriginalPrice      float64 `json:"originalPrice"`
	DiscountPercentage float64 `json:"discountPercentage"`
	StockQuantity      int     `json:"stockQuantity"`
	IsNew              bool    `json:"isNew"`
	IsBestSeller       bool    `json:"isBestSeller"`
	IsOnSale           bool    `json:"isOnSale"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedBy          string  `json:"created_by"`
	UpdatedBy          string  `json:"updated_by"`
	Status             string  `json:"status"`
}

func SeedData(session *gocql.Session, cfg *config.Config) error {
	// Check if products already exist
	var count int
	if err := session.Query(`SELECT COUNT(*) FROM products WHERE org = ? ALLOW FILTERING`,
		cfg.App.Org).Scan(&count); err != nil {
		log.Printf("Could not check product count: %v", err)
	}

	if count > 0 {
		log.Printf("Products already seeded (%d records), skipping.", count)
		return nil
	}

	// Read seed file
	data, err := os.ReadFile("mockup/product_data.json")
	if err != nil {
		log.Printf("Could not read seed file mockup/product_data.json: %v", err)
		return nil
	}

	var products []seedProduct
	if err := json.Unmarshal(data, &products); err != nil {
		return err
	}

	inserted := 0
	for _, p := range products {
		// Change app_name from senahapi to rootrevolutionapi
		p.AppName = cfg.App.Name

		createdAt := time.Now()
		if p.CreatedAt != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", p.CreatedAt); err == nil {
				createdAt = t
			}
		}

		updatedAt := time.Time{}

		err := session.Query(`INSERT INTO products
			(org, id, app_name, name, description, category, image,
			 price, original_price, discount_pct, stock_qty,
			 is_new, is_best_seller, is_on_sale,
			 created_at, updated_at, created_by, updated_by, status)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			p.Org, p.ID, p.AppName, p.Name, p.Description, p.Category, p.Image,
			p.Price, p.OriginalPrice, p.DiscountPercentage, p.StockQuantity,
			p.IsNew, p.IsBestSeller, p.IsOnSale,
			createdAt, updatedAt, p.CreatedBy, p.UpdatedBy, p.Status,
		).Exec()

		if err != nil {
			log.Printf("Failed to seed product %d (%s): %v", p.ID, p.Name, err)
			continue
		}
		inserted++
	}

	log.Printf("Seeded %d products successfully.", inserted)
	return nil
}
