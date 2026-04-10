package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	appproduct "rootrevolution-api/internal/application/product"
)

type ProductHandler struct {
	svc *appproduct.Service
}

func NewProductHandler(svc *appproduct.Service) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// ListProducts godoc
// GET /backend_rootrevolution/api/products
// Query: ?category=<category>
func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
	category := c.Query("category", "")
	products, err := h.svc.ListProducts(category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve products",
		})
	}

	if products == nil {
		return c.JSON(fiber.Map{
			"data":  []interface{}{},
			"total": 0,
		})
	}

	return c.JSON(fiber.Map{
		"data":  products,
		"total": len(products),
	})
}

// GetProduct godoc
// GET /backend_rootrevolution/api/products/:id
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	p, err := h.svc.GetProduct(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve product",
		})
	}
	if p == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	return c.JSON(fiber.Map{"data": p})
}

// CreateProduct godoc
// POST /backend_rootrevolution/api/products
// Requires auth. Sends authorization email to owner.
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req appproduct.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product name is required",
		})
	}
	if req.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product price must be greater than 0",
		})
	}

	requestedBy, _ := c.Locals("userEmail").(string)

	token, err := h.svc.CreateProduct(req, requestedBy)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Product creation request submitted. An authorization email has been sent to the owner for approval.",
		"token":   token,
	})
}

// UpdateProduct godoc
// PUT /backend_rootrevolution/api/products/:id
// Requires auth. Sends authorization email to owner.
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var req appproduct.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	requestedBy, _ := c.Locals("userEmail").(string)

	token, err := h.svc.UpdateProduct(id, req, requestedBy)
	if err != nil {
		if err.Error() == "product "+strconv.Itoa(id)+" not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Product update request submitted. An authorization email has been sent to the owner for approval.",
		"token":   token,
	})
}

// DeleteProduct godoc
// DELETE /backend_rootrevolution/api/products/:id
// Requires auth. Sends authorization email to owner.
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	requestedBy, _ := c.Locals("userEmail").(string)

	token, err := h.svc.DeleteProduct(id, requestedBy)
	if err != nil {
		if err.Error() == "product "+strconv.Itoa(id)+" not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Product deletion request submitted. An authorization email has been sent to the owner for approval.",
		"token":   token,
	})
}

// ListPending godoc
// GET /backend_rootrevolution/api/products/pending
// Admin only — returns all pending updates waiting for authorization
func (h *ProductHandler) ListPending(c *fiber.Ctx) error {
	items, err := h.svc.ListPending()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve pending updates",
		})
	}

	if items == nil {
		return c.JSON(fiber.Map{"data": []interface{}{}, "total": 0})
	}

	return c.JSON(fiber.Map{"data": items, "total": len(items)})
}

// ForceAuthorize godoc
// POST /backend_rootrevolution/api/products/pending/:token/approve
// Admin only — approves a pending update even if the email link has expired
func (h *ProductHandler) ForceAuthorize(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing token"})
	}

	result, err := h.svc.ForceAuthorize(token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Product change approved successfully",
		"data":    result,
	})
}

// AuthorizeUpdate godoc
// GET /backend_rootrevolution/api/products/authorize/:token
// Public - called when owner clicks email link
func (h *ProductHandler) AuthorizeUpdate(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing authorization token")
	}

	result, err := h.svc.AuthorizeUpdate(token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(
			"<html><body style='font-family:Arial;text-align:center;padding:50px'>" +
				"<h2 style='color:#e53e3e'>Authorization Failed</h2>" +
				"<p>" + err.Error() + "</p>" +
				"</body></html>",
		)
	}

	c.Set("Content-Type", "text/html")
	return c.Status(fiber.StatusOK).SendString(
		"<html><body style='font-family:Arial;text-align:center;padding:50px;background:#f9f9f9'>" +
			"<div style='max-width:500px;margin:auto;background:#fff;padding:40px;border-radius:10px;box-shadow:0 2px 10px rgba(0,0,0,0.1)'>" +
			"<div style='color:#38a169;font-size:60px'>&#10003;</div>" +
			"<h2 style='color:#1a1a2e'>Change Authorized</h2>" +
			"<p>Product <strong>" + strconv.Itoa(result.ID) + "</strong> has been successfully updated.</p>" +
			"<p style='color:#666;font-size:14px'>Root Revolution Product Management System</p>" +
			"</div></body></html>",
	)
}
