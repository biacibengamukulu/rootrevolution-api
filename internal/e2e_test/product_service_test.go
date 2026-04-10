package e2e_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"rootrevolution-api/config"
	appproduct "rootrevolution-api/internal/application/product"
	appuser "rootrevolution-api/internal/application/user"
	"rootrevolution-api/internal/domain/pending"
	"rootrevolution-api/internal/domain/product"
	"rootrevolution-api/internal/domain/user"
	"rootrevolution-api/internal/infrastructure/dropbox"
	"rootrevolution-api/internal/infrastructure/email"
	httphandler "rootrevolution-api/internal/interfaces/http"
)

// ─── In-memory product repository ────────────────────────────────────────────

type memProductRepo struct {
	mu       sync.Mutex
	products map[string]*product.Product
}

func newMemProductRepo(seed ...product.Product) *memProductRepo {
	r := &memProductRepo{products: make(map[string]*product.Product)}
	for _, p := range seed {
		cp := p
		r.products[fmt.Sprintf("%s:%d", p.Org, p.ID)] = &cp
	}
	return r
}

func (r *memProductRepo) key(org string, id int) string { return fmt.Sprintf("%s:%d", org, id) }

func (r *memProductRepo) FindAll(org string) ([]product.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []product.Product
	for _, p := range r.products {
		if p.Org == org {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (r *memProductRepo) FindByID(org string, id int) (*product.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.products[r.key(org, id)]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *memProductRepo) FindByCategory(org, category string) ([]product.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []product.Product
	for _, p := range r.products {
		if p.Org == org && p.Category == category {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (r *memProductRepo) Save(p *product.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *p
	r.products[r.key(p.Org, p.ID)] = &cp
	return nil
}

func (r *memProductRepo) Update(p *product.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *p
	r.products[r.key(p.Org, p.ID)] = &cp
	return nil
}

func (r *memProductRepo) Delete(org string, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.products, r.key(org, id))
	return nil
}

func (r *memProductRepo) NextID(org string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	max := 20100
	for _, p := range r.products {
		if p.Org == org && p.ID > max {
			max = p.ID
		}
	}
	return max + 1, nil
}

func (r *memProductRepo) Exists(org string, id int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.products[r.key(org, id)]
	return ok, nil
}

// ─── In-memory pending repository ────────────────────────────────────────────

type memPendingRepo struct {
	mu      sync.Mutex
	records map[uuid.UUID]*pending.PendingUpdate
}

func newMemPendingRepo() *memPendingRepo {
	return &memPendingRepo{records: make(map[uuid.UUID]*pending.PendingUpdate)}
}

func (r *memPendingRepo) Save(p *pending.PendingUpdate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *p
	r.records[p.Token] = &cp
	return nil
}

func (r *memPendingRepo) FindByToken(token uuid.UUID) (*pending.PendingUpdate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.records[token]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *memPendingRepo) UpdateStatus(token uuid.UUID, status pending.UpdateStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.records[token]; ok {
		p.Status = status
	}
	return nil
}

func (r *memPendingRepo) FindAllPending(org string) ([]*pending.PendingUpdate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*pending.PendingUpdate
	for _, p := range r.records {
		if p.Org == org && p.Status == pending.StatusPending {
			cp := *p
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memPendingRepo) CleanExpired() error { return nil }

// ─── In-memory user repository ───────────────────────────────────────────────

type memUserRepo struct {
	mu    sync.Mutex
	users map[string]*user.User
}

func newMemUserRepo(seed ...*user.User) *memUserRepo {
	r := &memUserRepo{users: make(map[string]*user.User)}
	for _, u := range seed {
		cp := *u
		r.users[u.Email] = &cp
	}
	return r
}

func (r *memUserRepo) FindByEmail(e string) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[e]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (r *memUserRepo) FindByID(id uuid.UUID) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.ID == id {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *memUserRepo) FindAll() ([]user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []user.User
	for _, u := range r.users {
		out = append(out, *u)
	}
	return out, nil
}

func (r *memUserRepo) Save(u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *u
	r.users[u.Email] = &cp
	return nil
}

func (r *memUserRepo) Update(u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *u
	r.users[u.Email] = &cp
	return nil
}

func (r *memUserRepo) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, u := range r.users {
		if u.ID == id {
			delete(r.users, k)
			return nil
		}
	}
	return nil
}

func (r *memUserRepo) Exists(e string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.users[e]
	return ok, nil
}

// ─── Test helpers ─────────────────────────────────────────────────────────────

const testOrg = "C10201"
const testJWTSecret = "test-secret"

// buildApp wires up a full Fiber app with in-memory repos and a fake email server.
// Returns the app and a signed admin JWT token ready to use.
func buildApp(productRepo product.Repository, pendingRepo pending.Repository) (*fiber.App, string) {
	emailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	}))

	cfg := &config.Config{
		App:     config.AppConfig{Name: "testapp", Org: testOrg, BaseURL: "http://localhost"},
		JWT:     config.JWTConfig{Secret: testJWTSecret},
		Owner:   config.OwnerConfig{Email: "owner@test.com"},
		Email:   config.EmailConfig{BaseURL: emailServer.URL, Org: "TESTORG"},
		Dropbox: config.DropboxConfig{BaseURL: "http://localhost"},
	}

	adminUser := user.NewUser("Admin", "Test", "admin@test.com", "hashed", user.RoleAdmin)
	userSvc := appuser.NewService(newMemUserRepo(adminUser), cfg)
	productSvc := appproduct.NewService(productRepo, pendingRepo, dropbox.NewClient(cfg), email.NewClient(cfg), cfg)

	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	app.Use(recover.New())
	httphandler.SetupRoutes(app, productSvc, userSvc, cfg)

	claims := jwt.MapClaims{
		"user_id": adminUser.ID.String(),
		"email":   adminUser.Email,
		"role":    string(user.RoleAdmin),
		"name":    adminUser.Name,
		"surname": adminUser.Surname,
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"sub":     adminUser.Email,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(testJWTSecret))

	return app, signed
}

func doRequest(app *fiber.App, method, path, body, token string) *http.Response {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, _ := app.Test(req, 5000)
	return resp
}

func parseBody(resp *http.Response) map[string]interface{} {
	var m map[string]interface{}
	b, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(b, &m)
	return m
}

// ─── Seed data ────────────────────────────────────────────────────────────────

func seedProducts() []product.Product {
	return []product.Product{
		{
			Org: testOrg, ID: 20101, AppName: "testapp",
			Name: "Botanical Bath Crystals", Category: "Skincare and beauty",
			Price: 120, OriginalPrice: 120, StockQuantity: 50,
			IsNew: true, Status: "active", CreatedAt: time.Now(),
		},
		{
			Org: testOrg, ID: 20102, AppName: "testapp",
			Name: "100% Natural Shampoo", Category: "Hair and scaptcare",
			Price: 95, OriginalPrice: 95, StockQuantity: 50,
			IsNew: true, Status: "active", CreatedAt: time.Now(),
		},
		{
			Org: testOrg, ID: 20103, AppName: "testapp",
			Name: "120 Moringa Capsules", Category: "Herbal and natural supplements",
			Price: 150, OriginalPrice: 150, StockQuantity: 50,
			IsNew: true, IsBestSeller: true, Status: "active", CreatedAt: time.Now(),
		},
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestListProducts_ReturnsAllProducts(t *testing.T) {
	app, _ := buildApp(newMemProductRepo(seedProducts()...), newMemPendingRepo())

	resp := doRequest(app, "GET", "/backend_rootrevolution/api/products", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := parseBody(resp)
	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatalf("expected 'data' array in response, got: %v", body)
	}
	if len(data) != 3 {
		t.Fatalf("expected 3 products, got %d", len(data))
	}
	if int(body["total"].(float64)) != 3 {
		t.Fatalf("expected total=3, got %v", body["total"])
	}
	t.Logf("ListProducts returned %d products", len(data))
}

func TestListProducts_FilterByCategory(t *testing.T) {
	app, _ := buildApp(newMemProductRepo(seedProducts()...), newMemPendingRepo())

	resp := doRequest(app, "GET", "/backend_rootrevolution/api/products?category=Skincare+and+beauty", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := parseBody(resp)
	data := body["data"].([]interface{})
	if len(data) != 1 {
		t.Fatalf("expected 1 skincare product, got %d", len(data))
	}
	name := data[0].(map[string]interface{})["name"].(string)
	if name != "Botanical Bath Crystals" {
		t.Fatalf("expected 'Botanical Bath Crystals', got %q", name)
	}
	t.Logf("Filter by category returned: %s", name)
}

func TestGetProduct_ByID(t *testing.T) {
	app, _ := buildApp(newMemProductRepo(seedProducts()...), newMemPendingRepo())

	resp := doRequest(app, "GET", "/backend_rootrevolution/api/products/20101", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := parseBody(resp)
	data := body["data"].(map[string]interface{})
	if data["name"] != "Botanical Bath Crystals" {
		t.Fatalf("expected 'Botanical Bath Crystals', got %v", data["name"])
	}
	t.Logf("GetProduct returned: %v", data["name"])
}

func TestGetProduct_NotFound(t *testing.T) {
	app, _ := buildApp(newMemProductRepo(seedProducts()...), newMemPendingRepo())

	resp := doRequest(app, "GET", "/backend_rootrevolution/api/products/99999", "", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	t.Log("GetProduct correctly returned 404 for unknown ID")
}

func TestListProducts_EmptyRepo(t *testing.T) {
	app, _ := buildApp(newMemProductRepo(), newMemPendingRepo())

	resp := doRequest(app, "GET", "/backend_rootrevolution/api/products", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := parseBody(resp)
	data := body["data"].([]interface{})
	if len(data) != 0 {
		t.Fatalf("expected empty array, got %d items", len(data))
	}
	t.Log("Empty repo returns empty data array")
}

func TestUpdateProduct_FullFlow(t *testing.T) {
	productRepo := newMemProductRepo(seedProducts()...)
	pendingRepo := newMemPendingRepo()
	app, token := buildApp(productRepo, pendingRepo)

	// Step 1: submit update — returns a pending auth token
	updateBody, _ := json.Marshal(map[string]interface{}{
		"name":  "Botanical Bath Crystals - Premium",
		"price": 135.0,
	})
	resp := doRequest(app, "PUT", "/backend_rootrevolution/api/products/20101", string(updateBody), token)
	if resp.StatusCode != http.StatusAccepted {
		body := parseBody(resp)
		t.Fatalf("expected 202, got %d — %v", resp.StatusCode, body)
	}

	body := parseBody(resp)
	authToken, ok := body["token"].(string)
	if !ok || authToken == "" {
		t.Fatalf("expected token in response, got: %v", body)
	}
	t.Logf("Update submitted, auth token: %s", authToken)

	// Step 2: owner clicks the authorization link
	authResp := doRequest(app, "GET", "/backend_rootrevolution/api/products/authorize/"+authToken, "", "")
	if authResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(authResp.Body)
		t.Fatalf("expected 200 on authorize, got %d — %s", authResp.StatusCode, string(b))
	}
	t.Log("Authorization successful")

	// Step 3: verify product was actually updated
	updated, err := productRepo.FindByID(testOrg, 20101)
	if err != nil || updated == nil {
		t.Fatalf("could not fetch updated product: %v", err)
	}
	if updated.Name != "Botanical Bath Crystals - Premium" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}
	if updated.Price != 135.0 {
		t.Fatalf("expected price 135.0, got %v", updated.Price)
	}
	t.Logf("Product updated successfully: name=%q price=%.2f", updated.Name, updated.Price)
}

func TestUpdateProduct_RequiresAuth(t *testing.T) {
	app, _ := buildApp(newMemProductRepo(seedProducts()...), newMemPendingRepo())

	body, _ := json.Marshal(map[string]interface{}{"name": "No Auth Update"})
	resp := doRequest(app, "PUT", "/backend_rootrevolution/api/products/20101", string(body), "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	t.Log("Update correctly blocked without auth token")
}

func TestListPending_ReturnsWaitingUpdates(t *testing.T) {
	productRepo := newMemProductRepo(seedProducts()...)
	pendingRepo := newMemPendingRepo()
	app, token := buildApp(productRepo, pendingRepo)

	// Submit two updates to build up pending records
	for _, id := range []string{"20101", "20102"} {
		body, _ := json.Marshal(map[string]interface{}{"price": 99.0})
		doRequest(app, "PUT", "/backend_rootrevolution/api/products/"+id, string(body), token)
	}

	resp := doRequest(app, "GET", "/backend_rootrevolution/api/products/pending", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := parseBody(resp)
	data := body["data"].([]interface{})
	if len(data) != 2 {
		t.Fatalf("expected 2 pending items, got %d", len(data))
	}
	t.Logf("ListPending returned %d pending updates", len(data))
}

func TestListPending_RequiresAdmin(t *testing.T) {
	app, _ := buildApp(newMemProductRepo(seedProducts()...), newMemPendingRepo())

	resp := doRequest(app, "GET", "/backend_rootrevolution/api/products/pending", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	t.Log("ListPending correctly blocked without auth")
}

func TestForceAuthorize_ApprovesExpiredToken(t *testing.T) {
	productRepo := newMemProductRepo(seedProducts()...)
	pendingRepo := newMemPendingRepo()
	app, token := buildApp(productRepo, pendingRepo)

	// Submit update
	updateBody, _ := json.Marshal(map[string]interface{}{"price": 200.0})
	resp := doRequest(app, "PUT", "/backend_rootrevolution/api/products/20103", string(updateBody), token)
	authToken := parseBody(resp)["token"].(string)

	// Simulate expiry by backdating the pending record
	parsedToken, _ := uuid.Parse(authToken)
	rec, _ := pendingRepo.FindByToken(parsedToken)
	rec.ExpiresAt = time.Now().Add(-1 * time.Hour)
	pendingRepo.records[parsedToken] = rec

	// Normal authorize should fail (expired)
	expiredResp := doRequest(app, "GET", "/backend_rootrevolution/api/products/authorize/"+authToken, "", "")
	if expiredResp.StatusCode == http.StatusOK {
		t.Fatal("expected authorize to fail on expired token, but it succeeded")
	}
	t.Log("Expired token correctly rejected by normal authorize")

	// Force approve should succeed
	approveResp := doRequest(app, "POST", "/backend_rootrevolution/api/products/pending/"+authToken+"/approve", "", token)
	if approveResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(approveResp.Body)
		t.Fatalf("expected 200 on force approve, got %d — %s", approveResp.StatusCode, string(b))
	}

	// Verify product was updated
	updated, _ := productRepo.FindByID(testOrg, 20103)
	if updated.Price != 200.0 {
		t.Fatalf("expected price 200.0 after force approve, got %v", updated.Price)
	}
	t.Logf("ForceAuthorize approved expired token, product price updated to %.2f", updated.Price)
}

func TestUpdateProduct_NotFound(t *testing.T) {
	app, token := buildApp(newMemProductRepo(seedProducts()...), newMemPendingRepo())

	body, _ := json.Marshal(map[string]interface{}{"name": "Ghost Product"})
	resp := doRequest(app, "PUT", "/backend_rootrevolution/api/products/99999", string(body), token)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	t.Log("Update correctly returned 404 for unknown product")
}
