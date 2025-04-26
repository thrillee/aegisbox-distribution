package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thrillee/aegisbox/internal/database"
)

// SetupRoutes configures the Gin engine with all API routes.
func SetupRoutes(router gin.IRouter, q database.Querier, pool *pgxpool.Pool) {
	// Initialize handlers
	spHandler := NewServiceProviderHandler(q, pool)
	reportHandler := NewReportHandler(q)

	// mnoHandler := NewMNOHandler(q, pool)
	// ... other handlers ...

	// --- Middleware (Example Placeholder) ---
	// authMiddleware := NewAuthMiddleware(...)
	// router.Use(authMiddleware)

	// --- Service Provider Routes ---
	spGroup := router.Group("/service-providers")
	{
		spGroup.POST("", spHandler.CreateServiceProvider)
		spGroup.GET("", spHandler.ListServiceProviders)
		spGroup.GET("/:id", spHandler.GetServiceProvider)
		// spGroup.PUT("/:id", spHandler.UpdateServiceProvider)
		// spGroup.DELETE("/:id", spHandler.DeleteServiceProvider)
	}

	// --- Reporting Routes ---
	reportGroup := router.Group("/reports")
	{
		reportGroup.GET("/daily", reportHandler.GetDailyReport)
	}

	// --- SP Credential Routes ---
	// POST /sp-credentials
	// GET /sp-credentials?sp_id={id}
	// PUT /sp-credentials/{id}
	// ...

	// --- MNO Routes ---
	// POST /mnos
	// GET /mnos
	// ...

	// --- MNO Connection Routes ---
	// POST /mno-connections
	// GET /mno-connections?mno_id={id}
	// ...

	// --- Routing Rule Routes ---
	// POST /routing-rules
	// GET /routing-rules
	// DELETE /routing-rules/{id}
	// ...

	// --- Sender ID Routes ---
	// POST /sender-ids
	// GET /sender-ids?sp_id={id}
	// PUT /sender-ids/{id}/status
	// ...

	// --- Pricing Rule Routes ---
	// POST /pricing-rules
	// GET /pricing-rules?sp_id={id}
	// DELETE /pricing-rules/{id}
	// ...

	// --- Wallet Routes ---
	// GET /wallets?sp_id={id}
	// POST /wallets/{id}/credit // Manual credit endpoint
	// ...
}
