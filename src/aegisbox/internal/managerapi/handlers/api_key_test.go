package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIKeyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func(key string) *gin.Engine {
		r := gin.New()
		v1 := r.Group("/api/v1")
		v1.Use(APIKeyMiddleware(key))
		ok := func(c *gin.Context) { c.Status(http.StatusOK) }
		v1.POST("/wallets/1/credit", ok)
		v1.POST("/webhooks/dlr/generic", ok)
		return r
	}

	tests := []struct {
		name    string
		key     string
		path    string
		headers map[string]string
		want    int
	}{
		{"no key configured allows", "", "/api/v1/wallets/1/credit", nil, http.StatusOK},
		{"missing key rejected", "secret", "/api/v1/wallets/1/credit", nil, http.StatusUnauthorized},
		{"wrong key rejected", "secret", "/api/v1/wallets/1/credit", map[string]string{"X-API-Key": "nope"}, http.StatusUnauthorized},
		{"X-API-Key accepted", "secret", "/api/v1/wallets/1/credit", map[string]string{"X-API-Key": "secret"}, http.StatusOK},
		{"bearer accepted", "secret", "/api/v1/wallets/1/credit", map[string]string{"Authorization": "Bearer secret"}, http.StatusOK},
		{"webhooks stay open", "secret", "/api/v1/webhooks/dlr/generic", nil, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			newRouter(tt.key).ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("got %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
