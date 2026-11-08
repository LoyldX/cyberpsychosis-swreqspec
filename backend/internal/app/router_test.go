package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestNewRouterBuilds ยืนยันว่าสร้าง router ได้และตอบ 404 กับ path ที่ยังไม่มี (CON-TECH-01; T-01).
func TestNewRouterBuilds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := NewRouter()
	if r == nil {
		t.Fatal("NewRouter returned nil")
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/not-exist", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown route, got %d", w.Code)
	}
}
