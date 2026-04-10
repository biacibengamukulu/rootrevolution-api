package http

import (
	"github.com/gofiber/fiber/v2"

	"rootrevolution-api/config"
	appproduct "rootrevolution-api/internal/application/product"
	appuser "rootrevolution-api/internal/application/user"
	"rootrevolution-api/internal/interfaces/middleware"
)

func SetupRoutes(
	app *fiber.App,
	productSvc *appproduct.Service,
	userSvc *appuser.Service,
	cfg *config.Config,
) {
	productHandler := NewProductHandler(productSvc)
	authHandler := NewAuthHandler(userSvc)
	authMiddleware := middleware.NewAuthMiddleware(userSvc)

	// API group with prefix
	api := app.Group("/backend_rootrevolution/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"app":     cfg.App.Name,
			"version": "1.0.0",
		})
	})

	// ─── Auth routes ────────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Get("/me", authMiddleware.Required(), authHandler.Me)
	auth.Post("/register", authMiddleware.Required(), authMiddleware.AdminOnly(), authHandler.Register)

	// ─── Product routes ─────────────────────────────────────────────────────────
	products := api.Group("/products")

	// Public - product authorization link (must be before /:id to avoid conflict)
	products.Get("/authorize/:token", productHandler.AuthorizeUpdate)

	// Admin only - pending updates management
	products.Get("/pending", authMiddleware.Required(), authMiddleware.AdminOnly(), productHandler.ListPending)
	products.Post("/pending/:token/approve", authMiddleware.Required(), authMiddleware.AdminOnly(), productHandler.ForceAuthorize)

	// Public - read operations
	products.Get("/", productHandler.ListProducts)
	products.Get("/:id", productHandler.GetProduct)

	// Protected - write operations
	products.Post("/", authMiddleware.Required(), productHandler.CreateProduct)
	products.Put("/:id", authMiddleware.Required(), productHandler.UpdateProduct)
	products.Delete("/:id", authMiddleware.Required(), productHandler.DeleteProduct)

	// ─── User management routes ─────────────────────────────────────────────────
	users := api.Group("/users", authMiddleware.Required(), authMiddleware.AdminOnly())
	users.Get("/", authHandler.ListUsers)
	users.Get("/:id", authHandler.GetUser)
	users.Put("/:id", authHandler.UpdateUser)
	users.Delete("/:id", authHandler.DeleteUser)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	})
}
