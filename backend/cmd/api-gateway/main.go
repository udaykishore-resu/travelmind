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
	"github.com/udaykishore-resu/travelmind/internal/config"
	"github.com/udaykishore-resu/travelmind/internal/db"
	"github.com/udaykishore-resu/travelmind/internal/handlers"
	"github.com/udaykishore-resu/travelmind/internal/middleware"
	"github.com/udaykishore-resu/travelmind/internal/observability"
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

	// Initialize tracer
	tracer, err := observability.InitTracer(cfg.Environment)
	if err != nil {
		logger.Fatalf("Failed to initialize tracer: %v", err)
	}

	// Initialize metrics
	metrics := observability.InitMetrics()

	// Initialize database
	pgDB, err := db.NewPostgresConnection(cfg.Database.PostgresURL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgDB.Close()

	// Initialize Redis
	redisClient := db.NewRedisConnection(cfg.Redis.URL)
	defer redisClient.Close()

	// Initialize Gin router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.TraceID())
	router.Use(middleware.CORS())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimiter())
	router.Use(middleware.ErrorHandling())
	router.Use(middleware.Recovery(logger))

	// Health check endpoints
	router.GET("/health", handlers.HealthCheck(pgDB, redisClient))
	router.GET("/ready", handlers.ReadinessCheck(pgDB, redisClient))
	router.GET("/metrics", gin.WrapH(metrics.Handler()))

	// API v1 routes
	h := handlers.New(pgDB, redisClient)
	v1 := router.Group("/api/v1")
	{
		// Auth endpoints
		v1.POST("/auth/register", h.Register)
		v1.POST("/auth/login", h.Login)
		v1.POST("/auth/refresh", h.RefreshToken)
		v1.POST("/auth/logout", middleware.AuthRequired(), h.Logout)

		// Traveler endpoints
		travelerGroup := v1.Group("/travelers")
		travelerGroup.Use(middleware.AuthRequired())
		{
			travelerGroup.GET("/:id", h.GetTraveler)
			travelerGroup.PUT("/:id", h.UpdateTraveler)
			travelerGroup.GET("/:id/preferences", h.GetTravelerPreferences)
			travelerGroup.PUT("/:id/preferences", h.UpdateTravelerPreferences)
		}

		// Booking endpoints
		bookingGroup := v1.Group("/bookings")
		bookingGroup.Use(middleware.AuthRequired())
		{
			bookingGroup.POST("", h.CreateBooking)
			bookingGroup.GET("/:id", h.GetBooking)
			bookingGroup.PUT("/:id", h.UpdateBooking)
			bookingGroup.DELETE("/:id", h.CancelBooking)
			bookingGroup.GET("", h.ListBookings)
			bookingGroup.POST("/:id/confirm", h.ConfirmBooking)
			bookingGroup.POST("/:id/pay", h.ProcessPayment)
		}

		// Supplier endpoints
		supplierGroup := v1.Group("/suppliers")
		{
			supplierGroup.GET("", h.ListSuppliers)
			supplierGroup.GET("/:id/rates", h.GetSupplierRates)
			supplierGroup.GET("/:id/availability", h.CheckSupplierAvailability)
		}

		// Search endpoints
		searchGroup := v1.Group("/search")
		{
			searchGroup.GET("/flights", h.SearchFlights)
			searchGroup.GET("/hotels", h.SearchHotels)
			searchGroup.GET("/activities", h.SearchActivities)
		}

		// AI endpoints
		aiGroup := v1.Group("/ai")
		aiGroup.Use(middleware.AuthRequired())
		{
			aiGroup.POST("/chat", h.AIChat)
			aiGroup.POST("/recommendations", h.GetRecommendations)
			aiGroup.POST("/risk-assessment", h.RiskAssessment)
		}

		// Advisor endpoints
		advisorGroup := v1.Group("/advisors")
		advisorGroup.Use(middleware.AuthRequired())
		{
			advisorGroup.GET("/:id", h.GetAdvisor)
			advisorGroup.PUT("/:id", h.UpdateAdvisor)
			advisorGroup.GET("/:id/bookings", h.GetAdvisorBookings)
			advisorGroup.GET("/:id/performance", h.GetAdvisorPerformance)
		}

		// Admin endpoints
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.AuthRequired())
		adminGroup.Use(middleware.RoleRequired("admin"))
		{
			adminGroup.GET("/users", h.ListUsers)
			adminGroup.GET("/bookings/analytics", h.GetBookingAnalytics)
			adminGroup.GET("/fraud/alerts", h.ListFraudAlerts)
			adminGroup.POST("/fraud/alerts/:id/resolve", h.ResolveFraudAlert)
		}
	}

	// Swagger/OpenAPI endpoint
	router.GET("/swagger", handlers.SwaggerUI())

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Infof("Starting API server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	// Cleanup
	tracer.Shutdown(ctx)
	metrics.Shutdown(ctx)

	logger.Info("Server stopped")
}
