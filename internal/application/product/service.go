package product

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"rootrevolution-api/config"
	"rootrevolution-api/internal/domain/pending"
	"rootrevolution-api/internal/domain/product"
	"rootrevolution-api/internal/infrastructure/dropbox"
	"rootrevolution-api/internal/infrastructure/email"
)

type CreateProductRequest struct {
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	Category           string  `json:"category"`
	Image              string  `json:"image"`
	Price              float64 `json:"price"`
	OriginalPrice      float64 `json:"original_price"`
	DiscountPercentage float64 `json:"discount_percentage"`
	StockQuantity      int     `json:"stock_quantity"`
	IsNew              bool    `json:"is_new"`
	IsBestSeller       bool    `json:"is_best_seller"`
	IsOnSale           bool    `json:"is_on_sale"`
	Status             string  `json:"status"`
}

type UpdateProductRequest struct {
	Name               *string  `json:"name"`
	Description        *string  `json:"description"`
	Category           *string  `json:"category"`
	Image              *string  `json:"image"`
	Price              *float64 `json:"price"`
	OriginalPrice      *float64 `json:"original_price"`
	DiscountPercentage *float64 `json:"discount_percentage"`
	StockQuantity      *int     `json:"stock_quantity"`
	IsNew              *bool    `json:"is_new"`
	IsBestSeller       *bool    `json:"is_best_seller"`
	IsOnSale           *bool    `json:"is_on_sale"`
	Status             *string  `json:"status"`
}

type Service struct {
	productRepo product.Repository
	pendingRepo pending.Repository
	dropbox     *dropbox.Client
	email       *email.Client
	cfg         *config.Config
}

func NewService(
	productRepo product.Repository,
	pendingRepo pending.Repository,
	dropboxClient *dropbox.Client,
	emailClient *email.Client,
	cfg *config.Config,
) *Service {
	return &Service{
		productRepo: productRepo,
		pendingRepo: pendingRepo,
		dropbox:     dropboxClient,
		email:       emailClient,
		cfg:         cfg,
	}
}

func (s *Service) ListProducts(category string) ([]product.Product, error) {
	if category != "" {
		return s.productRepo.FindByCategory(s.cfg.App.Org, category)
	}
	return s.productRepo.FindAll(s.cfg.App.Org)
}

func (s *Service) GetProduct(id int) (*product.Product, error) {
	return s.productRepo.FindByID(s.cfg.App.Org, id)
}

// CreateProduct stores a pending create and sends authorization email to owner
func (s *Service) CreateProduct(req CreateProductRequest, requestedBy string) (string, error) {
	nextID, err := s.productRepo.NextID(s.cfg.App.Org)
	if err != nil {
		return "", fmt.Errorf("generating product ID: %w", err)
	}

	imageURL := req.Image
	if req.Image != "" && dropbox.IsBase64Image(req.Image) {
		uploaded, err := s.dropbox.UploadBase64Image(req.Image, fmt.Sprintf("%d", nextID), "")
		if err != nil {
			return "", fmt.Errorf("uploading product image: %w", err)
		}
		imageURL = uploaded
	}

	status := "active"
	if req.Status != "" {
		status = req.Status
	}

	p := &product.Product{
		AppName:            s.cfg.App.Name,
		Org:                s.cfg.App.Org,
		ID:                 nextID,
		Name:               req.Name,
		Description:        req.Description,
		Category:           req.Category,
		Image:              imageURL,
		Price:              req.Price,
		OriginalPrice:      req.OriginalPrice,
		DiscountPercentage: req.DiscountPercentage,
		StockQuantity:      req.StockQuantity,
		IsNew:              req.IsNew,
		IsBestSeller:       req.IsBestSeller,
		IsOnSale:           req.IsOnSale,
		Status:             status,
		CreatedAt:          time.Now(),
		CreatedBy:          requestedBy,
	}

	data, _ := json.Marshal(p)
	pu := pending.NewPendingUpdate(nextID, s.cfg.App.Org, string(data), requestedBy, "create")

	if err := s.pendingRepo.Save(pu); err != nil {
		return "", fmt.Errorf("saving pending update: %w", err)
	}

	authLink := s.buildAuthLink(pu.Token.String())

	if err := s.email.SendProductAuthorizationEmail(
		s.cfg.Owner.Email, "Admin", authLink, "create", req.Name, requestedBy,
	); err != nil {
		return "", fmt.Errorf("sending authorization email: %w", err)
	}

	return pu.Token.String(), nil
}

// UpdateProduct stores a pending update and sends authorization email to owner
func (s *Service) UpdateProduct(id int, req UpdateProductRequest, requestedBy string) (string, error) {
	existing, err := s.productRepo.FindByID(s.cfg.App.Org, id)
	if err != nil {
		return "", err
	}
	if existing == nil {
		return "", fmt.Errorf("product %d not found", id)
	}

	if req.Image != nil && *req.Image != "" && dropbox.IsBase64Image(*req.Image) {
		uploaded, err := s.dropbox.UploadBase64Image(*req.Image, fmt.Sprintf("%d", id), "")
		if err != nil {
			return "", fmt.Errorf("uploading product image: %w", err)
		}
		req.Image = &uploaded
	}

	updated := *existing
	applyUpdates(&updated, req, requestedBy)

	data, _ := json.Marshal(updated)
	pu := pending.NewPendingUpdate(id, s.cfg.App.Org, string(data), requestedBy, "update")

	if err := s.pendingRepo.Save(pu); err != nil {
		return "", fmt.Errorf("saving pending update: %w", err)
	}

	authLink := s.buildAuthLink(pu.Token.String())

	if err := s.email.SendProductAuthorizationEmail(
		s.cfg.Owner.Email, "Admin", authLink, "update", existing.Name, requestedBy,
	); err != nil {
		return "", fmt.Errorf("sending authorization email: %w", err)
	}

	return pu.Token.String(), nil
}

// DeleteProduct stores a pending delete and sends authorization email to owner
func (s *Service) DeleteProduct(id int, requestedBy string) (string, error) {
	existing, err := s.productRepo.FindByID(s.cfg.App.Org, id)
	if err != nil {
		return "", err
	}
	if existing == nil {
		return "", fmt.Errorf("product %d not found", id)
	}

	payload, _ := json.Marshal(map[string]interface{}{"id": id, "org": s.cfg.App.Org})
	pu := pending.NewPendingUpdate(id, s.cfg.App.Org, string(payload), requestedBy, "delete")

	if err := s.pendingRepo.Save(pu); err != nil {
		return "", fmt.Errorf("saving pending delete: %w", err)
	}

	authLink := s.buildAuthLink(pu.Token.String())

	if err := s.email.SendProductAuthorizationEmail(
		s.cfg.Owner.Email, "Admin", authLink, "delete", existing.Name, requestedBy,
	); err != nil {
		return "", fmt.Errorf("sending authorization email: %w", err)
	}

	return pu.Token.String(), nil
}

// AuthorizeUpdate applies the pending change once the owner clicks the email link
func (s *Service) AuthorizeUpdate(tokenStr string) (*product.Product, error) {
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization token")
	}

	pu, err := s.pendingRepo.FindByToken(token)
	if err != nil {
		return nil, fmt.Errorf("finding pending update: %w", err)
	}
	if pu == nil {
		return nil, fmt.Errorf("authorization token not found")
	}
	if !pu.IsValid() {
		if pu.IsExpired() {
			return nil, fmt.Errorf("authorization link has expired")
		}
		return nil, fmt.Errorf("authorization already used or invalid")
	}

	var result *product.Product

	switch pu.Action {
	case "create":
		var p product.Product
		if err := json.Unmarshal([]byte(pu.UpdateData), &p); err != nil {
			return nil, fmt.Errorf("parsing product data: %w", err)
		}
		if err := s.productRepo.Save(&p); err != nil {
			return nil, fmt.Errorf("creating product: %w", err)
		}
		result = &p

	case "update":
		var p product.Product
		if err := json.Unmarshal([]byte(pu.UpdateData), &p); err != nil {
			return nil, fmt.Errorf("parsing product data: %w", err)
		}
		if err := s.productRepo.Update(&p); err != nil {
			return nil, fmt.Errorf("updating product: %w", err)
		}
		result = &p

	case "delete":
		if err := s.productRepo.Delete(pu.Org, pu.ProductID); err != nil {
			return nil, fmt.Errorf("deleting product: %w", err)
		}
		result = &product.Product{ID: pu.ProductID, Org: pu.Org, Status: "deleted"}

	default:
		return nil, fmt.Errorf("unknown action: %s", pu.Action)
	}

	if err := s.pendingRepo.UpdateStatus(pu.Token, pending.StatusApproved); err != nil {
		return result, nil
	}

	return result, nil
}

func (s *Service) buildAuthLink(token string) string {
	return fmt.Sprintf("%s/backend_rootrevolution/api/products/authorize/%s",
		s.cfg.App.BaseURL, token)
}

func applyUpdates(p *product.Product, req UpdateProductRequest, updatedBy string) {
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Category != nil {
		p.Category = *req.Category
	}
	if req.Image != nil {
		p.Image = *req.Image
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.OriginalPrice != nil {
		p.OriginalPrice = *req.OriginalPrice
	}
	if req.DiscountPercentage != nil {
		p.DiscountPercentage = *req.DiscountPercentage
	}
	if req.StockQuantity != nil {
		p.StockQuantity = *req.StockQuantity
	}
	if req.IsNew != nil {
		p.IsNew = *req.IsNew
	}
	if req.IsBestSeller != nil {
		p.IsBestSeller = *req.IsBestSeller
	}
	if req.IsOnSale != nil {
		p.IsOnSale = *req.IsOnSale
	}
	if req.Status != nil {
		p.Status = *req.Status
	}
	p.UpdatedBy = updatedBy
	p.UpdatedAt = time.Now()
}
