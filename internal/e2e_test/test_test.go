package e2e_test

import (
	"fmt"
	"log"
	"rootrevolution-api/config"
	appproduct "rootrevolution-api/internal/application/product"
	"rootrevolution-api/internal/infrastructure/cassandra"
	"rootrevolution-api/internal/infrastructure/dropbox"
	"rootrevolution-api/internal/infrastructure/email"
	"testing"

	myGlobal "github.com/biangacila/luvungula-go/global"
)

func TestProducts(t *testing.T) {
	cfg := config.Load()
	session, err := cassandra.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	defer session.Close()
	log.Println("Cassandra connected successfully")

	// ─── Repositories ───────────────────────────────────────────────────────────
	productRepo := cassandra.NewProductRepository(session)
	//userRepo := cassandra.NewUserRepository(session)
	pendingRepo := cassandra.NewPendingRepository(session)

	// ─── Infrastructure clients ─────────────────────────────────────────────────
	dropboxClient := dropbox.NewClient(cfg)
	emailClient := email.NewClient(cfg)

	// ─── Application services ───────────────────────────────────────────────────
	productSvc := appproduct.NewService(productRepo, pendingRepo, dropboxClient, emailClient, cfg)
	//userSvc := appuser.NewService(userRepo, cfg)

	t.Run("ProductService", func(t *testing.T) {
		list, err := productSvc.ListProducts("")
		if err != nil {
			t.Errorf("Error listing products: %v", err)
		}
		if len(list) == 0 {
			t.Errorf("Expected products, got none")
		}
		myGlobal.DisplayObject("list: "+fmt.Sprintf("%v", len(list)), list)

		category := "Skincare and beauty"
		list, err = productSvc.ListProducts(category)
		if err != nil {
			t.Errorf("Error listing products: %v", err)
		}
		if len(list) == 0 {
		}
		myGlobal.DisplayObject("list2 : "+fmt.Sprintf("%v", len(list)), list)

	})
}
