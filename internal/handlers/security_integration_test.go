package handlers

import (
	"bytes"
	"easywms-demo-v3/internal/auth"
	"easywms-demo-v3/internal/models"
	"easywms-demo-v3/internal/testdb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestMutationReplayAndConflict(t *testing.T) {
	db := testdb.Open(t)
	p := models.Product{SKU: "P", Name: "Product"}
	l := models.Location{Code: "A"}
	db.Create(&p)
	db.Create(&l)
	router := gin.New()
	h := Handler{DB: db}
	router.POST("/api/receive", func(c *gin.Context) { c.Set("user_id", "tester"); c.Set("username", "tester") }, h.Mutation("receive"))
	key := uuid.NewString()
	body := `{"sku":"P","location_code":"A","qty":3}`
	send := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/api/receive", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		r := httptest.NewRecorder()
		router.ServeHTTP(r, req)
		return r
	}
	var wg sync.WaitGroup
	results := make(chan *httptest.ResponseRecorder, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- send(body) }()
	}
	wg.Wait()
	close(results)
	for r := range results {
		if r.Code != 200 {
			t.Fatalf("status %d %s", r.Code, r.Body.String())
		}
	}
	var inv models.Inventory
	db.First(&inv)
	if inv.Qty != 3 {
		t.Fatal("duplicate stock", inv.Qty)
	}
	if r := send(`{"sku":"P","location_code":"A","qty":4}`); r.Code != 409 {
		t.Fatal("expected conflict", r.Code)
	}
}
func TestAuthUsesCurrentAccount(t *testing.T) {
	db := testdb.Open(t)
	employee := models.Employee{Username: "user", Code: "U", Name: "User", Role: "WAREHOUSE", Active: true}
	db.Create(&employee)
	h := Handler{DB: db, JWTSecret: "test-secret"}
	token, _ := auth.Sign(h.JWTSecret, auth.Claims{UserID: employee.ID.String(), Username: employee.Username, Role: "ADMIN", Exp: auth.ExpiresIn(1)})
	router := gin.New()
	router.GET("/admin", h.Auth(), RequireRoles("ADMIN"), func(c *gin.Context) { c.Status(200) })
	call := func() int {
		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r := httptest.NewRecorder()
		router.ServeHTTP(r, req)
		return r.Code
	}
	if call() != 403 {
		t.Fatal("stale role granted access")
	}
	db.Model(&employee).Update("active", false)
	if call() != 401 {
		t.Fatal("disabled account still authorized")
	}
}
