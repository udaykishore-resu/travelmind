package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Traveler represents a customer who books travel
type Traveler struct {
	ID              string     `json:"id" db:"id"`
	Email           string     `json:"email" db:"email"`
	PhoneNumber     string     `json:"phone_number" db:"phone_number"`
	FirstName       string     `json:"first_name" db:"first_name"`
	LastName        string     `json:"last_name" db:"last_name"`
	DateOfBirth     *time.Time `json:"date_of_birth" db:"date_of_birth"`
	ProfileSegment  string     `json:"profile_segment" db:"profile_segment"` // luxury, budget, adventure, family
	RiskScore       int        `json:"risk_score" db:"risk_score"`
	KYCVerified     bool       `json:"kyc_verified" db:"kyc_verified"`
	LifetimeValue   float64    `json:"lifetime_value" db:"lifetime_value"`
	LastBookingDate *time.Time `json:"last_booking_date" db:"last_booking_date"`
	PreferencesJSON JSONB      `json:"preferences" db:"preferences"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// Advisor represents a travel advisor
type Advisor struct {
	ID                 string    `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	Name               string    `json:"name" db:"name"`
	Email              string    `json:"email" db:"email"`
	Specialization     string    `json:"specialization" db:"specialization"` // cruise, luxury, adventure, corporate
	AvailabilityStatus string    `json:"availability_status" db:"availability_status"`
	CommissionRate     float64   `json:"commission_rate" db:"commission_rate"`
	Rating             float64   `json:"rating" db:"rating"`
	TotalBookings      int64     `json:"total_bookings" db:"total_bookings"`
	MonthlyRevenue     float64   `json:"monthly_revenue" db:"monthly_revenue"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Booking represents a complete travel booking
type Booking struct {
	ID                string     `json:"id" db:"id"`
	TravelerID        string     `json:"traveler_id" db:"traveler_id"`
	AdvisorID         *string    `json:"advisor_id" db:"advisor_id"`
	Status            string     `json:"status" db:"status"` // draft, pending, confirmed, paid, cancelled, completed
	TotalPrice        float64    `json:"total_price" db:"total_price"`
	Commission        float64    `json:"commission" db:"commission"`
	Currency          string     `json:"currency" db:"currency"`
	TravelDate        time.Time  `json:"travel_date" db:"travel_date"`
	BookingWindowDays int        `json:"booking_window_days" db:"booking_window_days"`
	ItineraryJSON     JSONB      `json:"itinerary" db:"itinerary"`
	NotesJSON         JSONB      `json:"notes" db:"notes"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	ConfirmedAt       *time.Time `json:"confirmed_at" db:"confirmed_at"`
	CancelledAt       *time.Time `json:"cancelled_at" db:"cancelled_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// BookingItem represents a single component of a booking (flight, hotel, etc.)
type BookingItem struct {
	ID                   string    `json:"id" db:"id"`
	BookingID            string    `json:"booking_id" db:"booking_id"`
	SupplierID           string    `json:"supplier_id" db:"supplier_id"` // "amadeus_flight", "booking_hotel"
	SupplierReference    string    `json:"supplier_reference" db:"supplier_reference"`
	ItemType             string    `json:"item_type" db:"item_type"` // flight, hotel, activity, transfer
	Price                float64   `json:"price" db:"price"`
	Status               string    `json:"status" db:"status"` // pending, confirmed, failed, cancelled
	SupplierResponseJSON JSONB     `json:"supplier_response" db:"supplier_response"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// Payment represents payment information
type Payment struct {
	ID               string     `json:"id" db:"id"`
	BookingID        string     `json:"booking_id" db:"booking_id"`
	Amount           float64    `json:"amount" db:"amount"`
	Currency         string     `json:"currency" db:"currency"`
	PaymentMethod    string     `json:"payment_method" db:"payment_method"` // card, bank_transfer
	Status           string     `json:"status" db:"status"`                 // pending, processing, successful, failed, refunded
	GatewayReference string     `json:"gateway_reference" db:"gateway_reference"`
	RiskScore        int        `json:"risk_score" db:"risk_score"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	ProcessedAt      *time.Time `json:"processed_at" db:"processed_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// FraudAlert represents a detected fraud risk
type FraudAlert struct {
	ID                string     `json:"id" db:"id"`
	BookingID         *string    `json:"booking_id" db:"booking_id"`
	TravelerID        string     `json:"traveler_id" db:"traveler_id"`
	RiskCategory      string     `json:"risk_category" db:"risk_category"`
	RiskScore         int        `json:"risk_score" db:"risk_score"`
	IsBlocked         bool       `json:"is_blocked" db:"is_blocked"`
	AdvisorOverrideID *string    `json:"advisor_override_id" db:"advisor_override_id"`
	OverrideReason    *string    `json:"override_reason" db:"override_reason"`
	DetectionJSON     JSONB      `json:"detection" db:"detection"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt        *time.Time `json:"resolved_at" db:"resolved_at"`
}

// SupplierRate represents cached supplier pricing
type SupplierRate struct {
	ID           int64     `json:"id" db:"id"`
	SupplierID   string    `json:"supplier_id" db:"supplier_id"`
	ProductKey   string    `json:"product_key" db:"product_key"` // "NYC-LAX-2026-09-30"
	Price        float64   `json:"price" db:"price"`
	Availability int       `json:"availability" db:"availability"`
	Currency     string    `json:"currency" db:"currency"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	FetchedAt    time.Time `json:"fetched_at" db:"fetched_at"`
}

// AuditLog represents an action taken in the system
type AuditLog struct {
	ID          int64     `json:"id" db:"id"`
	EntityType  string    `json:"entity_type" db:"entity_type"`
	EntityID    string    `json:"entity_id" db:"entity_id"`
	Action      string    `json:"action" db:"action"` // create, update, delete
	UserID      string    `json:"user_id" db:"user_id"`
	ChangesJSON JSONB     `json:"changes" db:"changes"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Supplier represents a travel service provider
type Supplier struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	SupplierType   string    `json:"supplier_type" db:"supplier_type"` // airline, hotel, activity
	APIKey         string    `json:"-" db:"api_key"`
	CommissionRate float64   `json:"commission_rate" db:"commission_rate"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// JSONB is a custom type for PostgreSQL JSONB columns
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	result := make(map[string]interface{})
	err := json.Unmarshal(bytes, &result)
	*j = JSONB(result)
	return err
}

// CreateTravelerRequest is the request body for creating a traveler
type CreateTravelerRequest struct {
	Email       string     `json:"email" binding:"required,email"`
	PhoneNumber string     `json:"phone_number" binding:"required"`
	FirstName   string     `json:"first_name" binding:"required"`
	LastName    string     `json:"last_name" binding:"required"`
	DateOfBirth *time.Time `json:"date_of_birth"`
}

// CreateBookingRequest is the request body for creating a booking
type CreateBookingRequest struct {
	TravelerID string    `json:"traveler_id" binding:"required,uuid"`
	Itinerary  JSONB     `json:"itinerary" binding:"required"`
	TravelDate time.Time `json:"travel_date" binding:"required"`
	Notes      *JSONB    `json:"notes"`
}

// ConfirmBookingRequest is the request body for confirming a booking
type ConfirmBookingRequest struct {
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
	Terms3DSSecure  bool   `json:"terms_3d_secure" binding:"required"`
}

// NewTraveler creates a new Traveler instance
func NewTraveler(email, phoneNumber, firstName, lastName string) *Traveler {
	return &Traveler{
		ID:          uuid.New().String(),
		Email:       email,
		PhoneNumber: phoneNumber,
		FirstName:   firstName,
		LastName:    lastName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewBooking creates a new Booking instance
func NewBooking(travelerID string, itinerary JSONB, travelDate time.Time) *Booking {
	return &Booking{
		ID:            uuid.New().String(),
		TravelerID:    travelerID,
		Status:        "draft",
		TravelDate:    travelDate,
		ItineraryJSON: itinerary,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// NewPayment creates a new Payment instance
func NewPayment(bookingID string, amount float64, paymentMethod string) *Payment {
	return &Payment{
		ID:            uuid.New().String(),
		BookingID:     bookingID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
