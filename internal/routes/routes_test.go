package routes

import (
	"easywms-demo-v3/internal/handlers"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMissingAPIIsNotHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	Register(router, handlers.Handler{})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/api/missing-endpoint", nil))
	if response.Code != 404 || !strings.Contains(response.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected JSON 404, got %d %s", response.Code, response.Header().Get("Content-Type"))
	}
}
