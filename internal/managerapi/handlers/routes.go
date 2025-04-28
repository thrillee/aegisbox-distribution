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
	credHandler := NewSPCredentialHandler(q, pool)
	reportHandler := NewReportHandler(q)

	mnoHandler := NewMNOHandler(q, pool)
	connHandler := NewMNOConnectionHandler(q, pool)

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
		spGroup.PUT("/:id", spHandler.UpdateServiceProvider)
		spGroup.DELETE("/:id", spHandler.DeleteServiceProvider)
	}

	// --- Reporting Routes ---
	reportGroup := router.Group("/reports")
	{
		reportGroup.GET("/daily", reportHandler.GetDailyReport)
	}

	// --- SP Credential Routes ---
	credGroup := router.Group("/sp-credentials")
	{
		credGroup.POST("", credHandler.CreateSPCredential)
		credGroup.GET("", credHandler.ListSPCredentials) // Filter with ?sp_id=...
		credGroup.GET("/:id", credHandler.GetSPCredential)
		credGroup.PUT("/:id", credHandler.UpdateSPCredential)
		credGroup.DELETE("/:id", credHandler.DeleteSPCredential)
	}

	// --- MNO Routes ---
	mnoGroup := router.Group("/mnos")
	{
		mnoGroup.POST("", mnoHandler.CreateMNO)
		mnoGroup.GET("", mnoHandler.ListMNOs) // Paginated
		mnoGroup.GET("/:id", mnoHandler.GetMNO)
		mnoGroup.PUT("/:id", mnoHandler.UpdateMNO)
		mnoGroup.DELETE("/:id", mnoHandler.DeleteMNO)

		connNestedGroup := mnoGroup.Group("/:mno_id/connections")
		{
			connNestedGroup.POST("", connHandler.CreateMNOConnection) // Handler needs to get mno_id from path
			connNestedGroup.GET("", connHandler.ListMNOConnections)   // Handler needs to get mno_id from path
		}

	}

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
