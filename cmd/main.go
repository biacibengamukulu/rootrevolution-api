package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"rootrevolution-api/config"
	appproduct "rootrevolution-api/internal/application/product"
	appuser "rootrevolution-api/internal/application/user"
	"rootrevolution-api/internal/infrastructure/cassandra"
	"rootrevolution-api/internal/infrastructure/dropbox"
	"rootrevolution-api/internal/infrastructure/email"
	httphandler "rootrevolution-api/internal/interfaces/http"
)

func main() {
	// Load .env if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	// ─── Cassandra ──────────────────────────────────────────────────────────────
	log.Printf("Connecting to Cassandra at %v...", cfg.Cassandra.Hosts)
	session, err := cassandra.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	defer session.Close()
	log.Println("Cassandra connected successfully")

	// ─── Migrations ─────────────────────────────────────────────────────────────
	log.Println("Running database migrations...")
	if err := cassandra.RunMigrations(session); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// ─── Seed Data ──────────────────────────────────────────────────────────────
	log.Println("Checking seed data...")
	if err := cassandra.SeedData(session, cfg); err != nil {
		log.Printf("Seed warning: %v", err)
	}

	// ─── Repositories ───────────────────────────────────────────────────────────
	productRepo := cassandra.NewProductRepository(session)
	userRepo := cassandra.NewUserRepository(session)
	pendingRepo := cassandra.NewPendingRepository(session)

	// ─── Infrastructure clients ─────────────────────────────────────────────────
	dropboxClient := dropbox.NewClient(cfg)
	emailClient := email.NewClient(cfg)

	// ─── Application services ───────────────────────────────────────────────────
	productSvc := appproduct.NewService(productRepo, pendingRepo, dropboxClient, emailClient, cfg)
	userSvc := appuser.NewService(userRepo, cfg)

	// ─── Ensure default admin ───────────────────────────────────────────────────
	log.Println("Ensuring default admin user...")
	if err := userSvc.EnsureDefaultAdmin(); err != nil {
		log.Printf("Warning: could not create default admin: %v", err)
	}

	// ─── Fiber App ──────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "rootrevolutionapi",
		ErrorHandler: errorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Authorization,Accept",
	}))

	// ─── Routes ─────────────────────────────────────────────────────────────────
	httphandler.SetupRoutes(app, productSvc, userSvc, cfg)

	// ─── Start server ────────────────────────────────────────────────────────────
	log.Printf("rootrevolutionapi starting on port %s", cfg.Server.Port)
	log.Printf("Base URL: %s/backend_rootrevolution/api", cfg.App.BaseURL)
	log.Fatal(app.Listen(":" + cfg.Server.Port))
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
