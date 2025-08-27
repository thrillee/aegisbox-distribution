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
	routingHandler := NewRoutingRuleHandler(q, pool)

	pricingHandler := NewPricingRuleHandler(q, pool)
	walletHandler := NewWalletHandler(q, pool)

	altSenderHandler := NewOtpAlternativeSenderHandler(q)
	templateHandler := NewOtpMessageTemplateHandler(q)
	assignmentHandler := NewOtpSenderTemplateAssignmentHandler(q)

	// ... other handlers ...

	// --- Middleware (Example Placeholder) ---
	// authMiddleware := NewAuthMiddleware(...)
	// router.Use(authMiddleware)

	otpSenders := router.Group("/otp-alternative-senders")
	{
		otpSenders.POST("", altSenderHandler.Create)
		otpSenders.GET("", altSenderHandler.List)
		otpSenders.GET("/:id", altSenderHandler.GetByID)
		otpSenders.PUT("/:id", altSenderHandler.Update)
		otpSenders.DELETE("/:id", altSenderHandler.Delete)
		otpSenders.GET("/senders-with-templates", altSenderHandler.ListSendersWithTemplates)
	}

	otpTemplates := router.Group("/otp-message-templates")
	{
		otpTemplates.POST("", templateHandler.Create)
		otpTemplates.GET("", templateHandler.List)
		otpTemplates.GET("/:id", templateHandler.GetByID)
		otpTemplates.PUT("/:id", templateHandler.Update)
		otpTemplates.DELETE("/:id", templateHandler.Delete)
	}

	otpAssignments := router.Group("/otp-sender-template-assignments")
	{
		otpAssignments.POST("", assignmentHandler.Create)
		otpAssignments.GET("", assignmentHandler.List) // Add query params for filtering if needed
		otpAssignments.GET("/:id", assignmentHandler.GetByID)
		otpAssignments.PUT("/:id", assignmentHandler.Update)
		otpAssignments.DELETE("/:id", assignmentHandler.Delete)
	}

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

		connNestedGroup := mnoGroup.Group("/connections/:mno_id/")
		{
			connNestedGroup.POST("", connHandler.CreateMNOConnection) // Handler needs to get mno_id from path
			connNestedGroup.GET("", connHandler.ListMNOConnections)   // Handler needs to get mno_id from path
		}

	}

	// --- Routing Rule Routes ---
	ruleGroup := router.Group("/routing-rules")
	{
		ruleGroup.POST("", routingHandler.CreateRule)
		ruleGroup.GET("", routingHandler.ListRules) // Optional filter ?mno_id=...
		ruleGroup.GET("/:id", routingHandler.GetRule)
		ruleGroup.PUT("/:id", routingHandler.UpdateRule)
		ruleGroup.DELETE("/:id", routingHandler.DeleteRule)
	}

	// --- Sender ID Routes ---
	// POST /sender-ids
	// GET /sender-ids?sp_id={id}
	// PUT /sender-ids/{id}/status
	// ...

	// --- Pricing Rule Routes ---
	pricingGroup := router.Group("/pricing-rules")
	{
		pricingGroup.POST("", pricingHandler.CreatePricingRule)
		pricingGroup.GET("", pricingHandler.ListPricingRules) // Requires ?sp_id=... filter
		pricingGroup.DELETE("/:id", pricingHandler.DeletePricingRule)
	}

	// --- Wallet Routes ---
	walletGroup := router.Group("/wallets")
	{
		walletGroup.GET("", walletHandler.ListWallets) // Requires ?sp_id=... filter
		// GET /wallets/:id might be useful too
		// walletGroup.GET("/:id", walletHandler.GetWallet)
		walletGroup.POST("/:id/credit", walletHandler.CreditWallet) // Manual credit
		// DELETE /wallets/:id ? Probably not needed, deleted with SP.
		// PUT /wallets/:id ? Maybe to update low_balance_threshold?
	}
}
