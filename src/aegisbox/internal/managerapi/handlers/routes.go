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
	routeSenderHandler := NewRouteSenderHandler(q)
	routingHandler := NewRoutingAssignmentHandler(q, pool)

	pricingHandler := NewPricingRuleHandler(q, pool)
	walletHandler := NewWalletHandler(q, pool)

	altSenderHandler := NewOtpAlternativeSenderHandler(q)
	templateHandler := NewOtpMessageTemplateHandler(q)
	assignmentHandler := NewOtpSenderTemplateAssignmentHandler(q)

	// Health monitoring and routing config handlers
	healthBroadcaster := NewHealthEventBroadcaster()
	healthHandler := NewHealthHandler(q, pool, healthBroadcaster)
	configHandler := NewRoutingConfigHandler(q, pool)

	// Webhook handler for HTTP MNO DLRs
	webhookHandler := NewWebhookHandler(q)

	// Manual DLR and auto DLR config handler
	dlrHandler := NewDLRHandler(q, pool)

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
		spGroup.GET("/:id/auto-dlr", dlrHandler.GetAutoDLR)
		spGroup.PUT("/:id/auto-dlr", dlrHandler.UpdateAutoDLR)
		spGroup.GET("/:id/route-sender-pool", routeSenderHandler.GetRouteSenderPoolSetting)
		spGroup.PUT("/:id/route-sender-pool", routeSenderHandler.UpdateRouteSenderPoolSetting)
	}

	// --- Route Sender Pool Routes (sender IDs each upstream connection accepts) ---
	router.GET("/mno-connections/:id/route-senders", routeSenderHandler.GetRouteSenderPool)
	router.POST("/mno-connections/:id/route-senders", routeSenderHandler.CreateRouteSender)
	router.PUT("/mno-connections/:id/route-senders/strategy", routeSenderHandler.UpdateSenderPoolStrategy)
	router.PUT("/route-senders/:id", routeSenderHandler.UpdateRouteSender)
	router.DELETE("/route-senders/:id", routeSenderHandler.DeleteRouteSender)

	// --- Message DLR Routes ---
	messageGroup := router.Group("/messages")
	{
		messageGroup.POST("/:id/dlr", dlrHandler.SendManualDLR)      // Manual DLR for one message
		messageGroup.POST("/dlr/bulk", dlrHandler.SendBulkManualDLR) // Manual DLR for many messages
	}

	// --- DLR Forwarding Queue Routes ---
	router.POST("/dlr-forwarding/retry", dlrHandler.RetryDLRForwarding) // Requeue failed DLR forwarding jobs

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

	// --- Routing Assignment Routes ---
	routingGroup := router.Group("/routing-assignments")
	{
		routingGroup.POST("", routingHandler.Create)
		routingGroup.GET("", routingHandler.List)
		routingGroup.GET("/:id", routingHandler.GetByID)
		routingGroup.PUT("/:id", routingHandler.Update)
		routingGroup.DELETE("/:id", routingHandler.Delete)
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

	// --- Health Monitoring Routes ---
	healthGroup := router.Group("/routing/health")
	{
		healthGroup.POST("/routes/:assignment_id/mark-available", healthHandler.MarkRouteAvailable)
		healthGroup.POST("/routes/:assignment_id/mark-unavailable", healthHandler.MarkRouteAvailable)
		healthGroup.POST("/senders/:sender_id/:mno_id/unblock", healthHandler.UnblockSender)
		healthGroup.GET("/events/stream", healthHandler.StreamHealthEvents)
		healthGroup.GET("/stats", healthHandler.GetHealthStats)
		healthGroup.GET("/events/recent", healthHandler.GetRecentHealthEvents)
	}

	// --- Routing Configuration Routes ---
	configGroup := router.Group("/routing/config")
	{
		configGroup.GET("", configHandler.ListConfigs)
		configGroup.GET("/:group_id", configHandler.GetConfig)
		configGroup.POST("", configHandler.CreateConfig)
		configGroup.PUT("/:group_id", configHandler.UpdateConfig)
		configGroup.GET("/global", configHandler.GetGlobalConfig)
		configGroup.PUT("/global", configHandler.UpdateGlobalConfig)
	}

	// --- Webhook Routes for HTTP MNO DLRs ---
	// Apply security middleware to webhook routes
	securityConfig := DefaultWebhookSecurityConfig()
	securityMiddleware := NewWebhookSecurityMiddleware(securityConfig)

	webhookGroup := router.Group("/webhooks/dlr")
	webhookGroup.Use(securityMiddleware.Middleware())
	{
		// Provider-agnostic webhook endpoint
		webhookGroup.POST("/generic", webhookHandler.HandleProviderWebhook)

		// Provider-specific webhook endpoints (cleaner URLs for providers)
		webhookGroup.POST("/termii", webhookHandler.HandleTermiiWebhook)
		webhookGroup.POST("/africatalking", webhookHandler.HandleAfricaTalkingWebhook)
		webhookGroup.POST("/simpu", webhookHandler.HandleSimpuWebhook)
		webhookGroup.POST("/simpu-otp", webhookHandler.HandleSimpuOTPWebhook)
		webhookGroup.POST("/ebulksms", webhookHandler.HandleEBulkSMSWebhook)
		webhookGroup.POST("/multitexter", webhookHandler.HandleMultiTexterWebhook)
	}
}
