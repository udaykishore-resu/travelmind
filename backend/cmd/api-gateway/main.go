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
	"github.com/sirupsen/logrus"
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
	pgDB, err := db.NewPostgresConnection(cfg.Database.URL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgDB.Close()

	// Initialize Redis
	redisClient := db.NewRedisClient(cfg.Redis.URL)
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
	router.GET("/health", func(c *gin.Context) {
		handlers.HealthCheck(c, pgDB, redisClient)
	})
	router.GET("/ready", func(c *gin.Context) {
		handlers.ReadinessCheck(c, pgDB, redisClient)
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth endpoints
		v1.POST("/auth/register", handlers.Register)
		v1.POST("/auth/login", handlers.Login)
		v1.POST("/auth/refresh", handlers.RefreshToken)
		v1.POST("/auth/logout", middleware.AuthRequired(), handlers.Logout)

		// Traveler endpoints
		travelerGroup := v1.Group("/travelers")
		travelerGroup.Use(middleware.AuthRequired())
		{
			travelerGroup.GET("/:id", handlers.GetTraveler)
			travelerGroup.PUT("/:id", handlers.UpdateTraveler)
			travelerGroup.GET("/:id/preferences", handlers.GetTravelerPreferences)
			travelerGroup.PUT("/:id/preferences", handlers.UpdateTravelerPreferences)
		}

		// Booking endpoints
		bookingGroup := v1.Group("/bookings")
		bookingGroup.Use(middleware.AuthRequired())
		{
			bookingGroup.POST("", handlers.CreateBooking)
			bookingGroup.GET("/:id", handlers.GetBooking)
			bookingGroup.PUT("/:id", handlers.UpdateBooking)
			bookingGroup.DELETE("/:id", handlers.CancelBooking)
			bookingGroup.GET("", handlers.ListBookings)
			bookingGroup.POST("/:id/confirm", handlers.ConfirmBooking)
			bookingGroup.POST("/:id/pay", handlers.ProcessPayment)
		}

		// Supplier endpoints
		supplierGroup := v1.Group("/suppliers")
		{
			supplierGroup.GET("", handlers.ListSuppliers)
			supplierGroup.GET("/:id/rates", handlers.GetSupplierRates)
			supplierGroup.GET("/:id/availability", handlers.CheckSupplierAvailability)
		}

		// Search endpoints
		searchGroup := v1.Group("/search")
		{
			searchGroup.GET("/flights", handlers.SearchFlights)
			searchGroup.GET("/hotels", handlers.SearchHotels)
			searchGroup.GET("/activities", handlers.SearchActivities)
		}

		// AI endpoints
		aiGroup := v1.Group("/ai")
		aiGroup.Use(middleware.AuthRequired())
		{
			aiGroup.POST("/chat", handlers.AIChat)
			aiGroup.POST("/recommendations", handlers.GetRecommendations)
			aiGroup.POST("/risk-assessment", handlers.RiskAssessment)
		}

		// Advisor endpoints
		advisorGroup := v1.Group("/advisors")
		advisorGroup.Use(middleware.AuthRequired())
		{
			advisorGroup.GET("/:id", handlers.GetAdvisor)
			advisorGroup.PUT("/:id", handlers.UpdateAdvisor)
			advisorGroup.GET("/:id/bookings", handlers.GetAdvisorBookings)
			advisorGroup.GET("/:id/performance", handlers.GetAdvisorPerformance)
		}

		// Admin endpoints
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.AuthRequired())
		adminGroup.Use(middleware.RoleRequired("admin"))
		{
			adminGroup.GET("/users", handlers.ListUsers)
			adminGroup.GET("/bookings/analytics", handlers.GetBookingAnalytics)
			adminGroup.GET("/fraud/alerts", handlers.ListFraudAlerts)
			adminGroup.POST("/fraud/alerts/:id/resolve", handlers.ResolveFraudAlert)
		}
	}

	// Swagger/OpenAPI endpoint
	router.GET("/swagger", handlers.SwaggerUI)

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
