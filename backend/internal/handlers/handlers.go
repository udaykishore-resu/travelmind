package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/udaykishore-resu/travelmind/internal/db"
)

// Handler holds the shared dependencies for all API v1 endpoints.
type Handler struct {
	pg    *db.PostgresDB
	redis *redis.Client
}

// New creates a Handler wired to the given data stores.
func New(pg *db.PostgresDB, redisClient *redis.Client) *Handler {
	return &Handler{pg: pg, redis: redisClient}
}

// notImplemented responds with 501 so clients and tests can tell a stubbed
// endpoint apart from a working one.
func notImplemented(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":    "not implemented",
			"endpoint": name,
		})
	}
}

// Auth

func (h *Handler) Register(c *gin.Context)     { notImplemented("Register")(c) }
func (h *Handler) Login(c *gin.Context)        { notImplemented("Login")(c) }
func (h *Handler) RefreshToken(c *gin.Context) { notImplemented("RefreshToken")(c) }
func (h *Handler) Logout(c *gin.Context)       { notImplemented("Logout")(c) }

// Travelers

func (h *Handler) GetTraveler(c *gin.Context)    { notImplemented("GetTraveler")(c) }
func (h *Handler) UpdateTraveler(c *gin.Context) { notImplemented("UpdateTraveler")(c) }
func (h *Handler) GetTravelerPreferences(c *gin.Context) {
	notImplemented("GetTravelerPreferences")(c)
}
func (h *Handler) UpdateTravelerPreferences(c *gin.Context) {
	notImplemented("UpdateTravelerPreferences")(c)
}

// Bookings

func (h *Handler) CreateBooking(c *gin.Context)  { notImplemented("CreateBooking")(c) }
func (h *Handler) GetBooking(c *gin.Context)     { notImplemented("GetBooking")(c) }
func (h *Handler) UpdateBooking(c *gin.Context)  { notImplemented("UpdateBooking")(c) }
func (h *Handler) CancelBooking(c *gin.Context)  { notImplemented("CancelBooking")(c) }
func (h *Handler) ListBookings(c *gin.Context)   { notImplemented("ListBookings")(c) }
func (h *Handler) ConfirmBooking(c *gin.Context) { notImplemented("ConfirmBooking")(c) }
func (h *Handler) ProcessPayment(c *gin.Context) { notImplemented("ProcessPayment")(c) }

// Suppliers

func (h *Handler) ListSuppliers(c *gin.Context)    { notImplemented("ListSuppliers")(c) }
func (h *Handler) GetSupplierRates(c *gin.Context) { notImplemented("GetSupplierRates")(c) }
func (h *Handler) CheckSupplierAvailability(c *gin.Context) {
	notImplemented("CheckSupplierAvailability")(c)
}

// Search

func (h *Handler) SearchFlights(c *gin.Context)    { notImplemented("SearchFlights")(c) }
func (h *Handler) SearchHotels(c *gin.Context)     { notImplemented("SearchHotels")(c) }
func (h *Handler) SearchActivities(c *gin.Context) { notImplemented("SearchActivities")(c) }

// AI

func (h *Handler) AIChat(c *gin.Context)             { notImplemented("AIChat")(c) }
func (h *Handler) GetRecommendations(c *gin.Context) { notImplemented("GetRecommendations")(c) }
func (h *Handler) RiskAssessment(c *gin.Context)     { notImplemented("RiskAssessment")(c) }

// Advisors

func (h *Handler) GetAdvisor(c *gin.Context)         { notImplemented("GetAdvisor")(c) }
func (h *Handler) UpdateAdvisor(c *gin.Context)      { notImplemented("UpdateAdvisor")(c) }
func (h *Handler) GetAdvisorBookings(c *gin.Context) { notImplemented("GetAdvisorBookings")(c) }
func (h *Handler) GetAdvisorPerformance(c *gin.Context) {
	notImplemented("GetAdvisorPerformance")(c)
}

// Admin

func (h *Handler) ListUsers(c *gin.Context)           { notImplemented("ListUsers")(c) }
func (h *Handler) GetBookingAnalytics(c *gin.Context) { notImplemented("GetBookingAnalytics")(c) }
func (h *Handler) ListFraudAlerts(c *gin.Context)     { notImplemented("ListFraudAlerts")(c) }
func (h *Handler) ResolveFraudAlert(c *gin.Context)   { notImplemented("ResolveFraudAlert")(c) }
