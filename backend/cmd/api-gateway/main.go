package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/revansystems/travelmind/internal/config"
	"github.com/revansystems/travelmind/internal/db"
	"github.com/revansystems/travelmind/internal/handlers"
	"github.com/revansystems/travelmind/internal/middleware"
	"github.com/revansystems/travelmind/internal/observability"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := observability.NewLogger(cfg.LogLevel)
	logger.WithFields(logrus.Fields{
		"environment": cfg.Environment,
		"version":     cfg.Version,
	}).Info("TravelMind API Gateway starting")

	// Initialize observability (tracing, metrics)
	tracer, err := observability.InitTracer(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize tracer")
	}
	defer tracer.Shutdown(context.Background())

	metricsServer := observability.InitMetrics()
	go func() {
		if err := http.ListenAndServe(":9090", metricsServer); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("Metrics server failed")
		}
	}()

	// Initialize database connections
	pgDB, err := db.NewPostgresConnection(cfg.Database.PostgresURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to PostgreSQL")
	}
	defer pgDB.Close()

	redisClient := db.NewRedisConnection(cfg.Redis.URL)
	defer redisClient.Close()

	// Health checks
	if err := pgDB.Ping(context.Background()); err != nil {
		logger.WithError(err).Fatal("PostgreSQL health check failed")
	}
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.WithError(err).Fatal("Redis health check failed")
	}

	logger.Info("Database connections established")

	// Initialize Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middleware
	router.Use(
		gin.Recovery(),
		middleware.RequestLogger(logger),
		middleware.TraceID(),
		middleware.CORS(cfg),
		middleware.SecurityHeaders(),
	)

	// Health check endpoints (no auth required)
	router.GET("/health", handlers.HealthCheck(pgDB, redisClient))
	router.GET("/ready", handlers.ReadinessCheck(pgDB, redisClient))
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "Metrics available at :9090/metrics")
	})

	// API Documentation
	router.GET("/swagger", handlers.SwaggerUI())
	router.StaticFile("/swagger.yaml", "./docs/swagger.yaml")

	// Public routes (no authentication)
	public := router.Group("/api/v1")
	{
		public.POST("/auth/login", handlers.Login())
		public.POST("/auth/signup", handlers.Signup(pgDB))
		public.POST("/auth/refresh", handlers.RefreshToken())
	}

	// Protected routes (JWT authentication required)
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthRequired())
	{
		// Traveler endpoints
		travelers := protected.Group("/travelers")
		{
			travelers.GET("/:id", handlers.GetTraveler(pgDB))
			travelers.PUT("/:id", handlers.UpdateTraveler(pgDB))
			travelers.GET("/:id/bookings", handlers.GetTravelerBookings(pgDB))
			travelers.GET("/:id/preferences", handlers.GetTravelerPreferences(pgDB))
			travelers.PUT("/:id/preferences", handlers.UpdateTravelerPreferences(pgDB))
		}

		// Booking endpoints
		bookings := protected.Group("/bookings")
		{
			bookings.POST("", handlers.CreateBooking(pgDB, redisClient))
			bookings.GET("/:id", handlers.GetBooking(pgDB))
			bookings.PUT("/:id", handlers.UpdateBooking(pgDB))
			bookings.POST("/:id/confirm", handlers.ConfirmBooking(pgDB))
			bookings.POST("/:id/cancel", handlers.CancelBooking(pgDB))
			bookings.GET("/:id/items", handlers.GetBookingItems(pgDB))
		}

		// Supplier endpoints
		suppliers := protected.Group("/suppliers")
		{
			suppliers.GET("", handlers.ListSuppliers(pgDB))
			suppliers.GET("/:id", handlers.GetSupplier(pgDB))
			suppliers.POST("/:id/rates", handlers.GetSupplierRates(pgDB, redisClient))
		}

		// Search endpoints
		search := protected.Group("/search")
		{
			search.POST("/flights", handlers.SearchFlights(pgDB, redisClient))
			search.POST("/hotels", handlers.SearchHotels(pgDB, redisClient))
			search.POST("/activities", handlers.SearchActivities(pgDB, redisClient))
		}

		// AI Agent endpoints
		ai := protected.Group("/ai")
		{
			ai.POST("/chat", handlers.ChatWithAgent())
			ai.POST("/recommend", handlers.GetRecommendations(pgDB))
			ai.POST("/risk-assess", handlers.AssessBookingRisk(pgDB))
		}

		// Advisor endpoints (RBAC: advisor_role or higher)
		advisors := protected.Group("/advisors")
		advisors.Use(middleware.RoleRequired("advisor"))
		{
			advisors.GET("/dashboard", handlers.AdvisorDashboard(pgDB))
			advisors.GET("/clients", handlers.ListAdvisorClients(pgDB))
			advisors.GET("/:id/performance", handlers.GetAdvisorPerformance(pgDB))
			advisors.POST("/bookings/:id/override", handlers.OverrideFraudBlock(pgDB))
		}

		// Admin endpoints (RBAC: admin_role only)
		admin := protected.Group("/admin")
		admin.Use(middleware.RoleRequired("admin"))
		{
			admin.GET("/users", handlers.ListUsers(pgDB))
			admin.POST("/users/:id/role", handlers.UpdateUserRole(pgDB))
			admin.GET("/audit-logs", handlers.GetAuditLogs(pgDB))
			admin.GET("/system-health", handlers.SystemHealth(pgDB, redisClient))
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
			"path":  c.Request.URL.Path,
		})
	})

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.WithField("port", cfg.Server.Port).Info("API Gateway listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server error")
		}
	}()

	// Wait for interrupt signal
	<-quit
	logger.Info("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced shutdown")
	}

	logger.Info("Server exited successfully")
}
