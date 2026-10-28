package handlers

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// webhookPathPrefix is left open: MNOs post DLRs there and it has its own security middleware.
const webhookPathPrefix = "/api/v1/webhooks/"

// APIKeyMiddleware requires the X-API-Key header (or "Authorization: Bearer <key>") to match
// apiKey on every route except the MNO DLR webhooks. An empty apiKey disables the check.
func APIKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" || strings.HasPrefix(c.Request.URL.Path, webhookPathPrefix) {
			c.Next()
			return
		}

		provided := c.GetHeader("X-API-Key")
		if provided == "" {
			provided = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing API key"})
			return
		}
		c.Next()
	}
}
