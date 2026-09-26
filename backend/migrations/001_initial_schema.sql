-- TravelMind Database Schema
-- Version: 1.0
-- Description: Initial schema for travelers, advisors, bookings, and payments

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Travelers table
CREATE TABLE travelers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  phone_country VARCHAR(3),
  phone_number VARCHAR(20),
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  date_of_birth DATE,
  profile_segment VARCHAR(50), -- luxury, budget, adventure, family, business
  risk_score INTEGER DEFAULT 0,
  kyc_verified BOOLEAN DEFAULT FALSE,
  lifetime_value DECIMAL(10, 2) DEFAULT 0,
  last_booking_date TIMESTAMP WITH TIME ZONE,
  preferences JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_travelers_email ON travelers(email);
CREATE INDEX idx_travelers_risk_score ON travelers(risk_score);
CREATE INDEX idx_travelers_kyc ON travelers(kyc_verified);

-- Advisors table
CREATE TABLE advisors (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  phone_number VARCHAR(20),
  specialization VARCHAR(100), -- cruise, luxury, adventure, corporate
  availability_status VARCHAR(20) DEFAULT 'offline', -- available, busy, offline
  commission_rate DECIMAL(5, 2),
  rating DECIMAL(3, 2) DEFAULT 5.0,
  total_bookings INTEGER DEFAULT 0,
  monthly_revenue DECIMAL(10, 2) DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_advisors_email ON advisors(email);
CREATE INDEX idx_advisors_specialization ON advisors(specialization);
CREATE INDEX idx_advisors_availability ON advisors(availability_status);

-- Bookings table
CREATE TABLE bookings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  traveler_id UUID NOT NULL REFERENCES travelers(id) ON DELETE CASCADE,
  advisor_id UUID REFERENCES advisors(id) ON DELETE SET NULL,
  status VARCHAR(50) DEFAULT 'draft', -- draft, pending, confirmed, paid, cancelled, completed
  total_price DECIMAL(10, 2),
  commission DECIMAL(10, 2),
  currency VARCHAR(3) DEFAULT 'USD',
  travel_date DATE NOT NULL,
  booking_window_days INTEGER,
  itinerary JSONB DEFAULT '{}'::jsonb,
  notes JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  confirmed_at TIMESTAMP WITH TIME ZONE,
  cancelled_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bookings_traveler ON bookings(traveler_id);
CREATE INDEX idx_bookings_advisor ON bookings(advisor_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_travel_date ON bookings(travel_date);
CREATE INDEX idx_bookings_status_date ON bookings(status, travel_date);

-- Booking items table
CREATE TABLE booking_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
  supplier_id VARCHAR(100) NOT NULL,
  supplier_reference VARCHAR(255),
  item_type VARCHAR(50), -- flight, hotel, activity, transfer
  price DECIMAL(10, 2),
  status VARCHAR(50) DEFAULT 'pending', -- pending, confirmed, failed, cancelled
  supplier_response JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_booking_items_booking ON booking_items(booking_id);
CREATE INDEX idx_booking_items_supplier ON booking_items(supplier_id);
CREATE INDEX idx_booking_items_status ON booking_items(status);

-- Payments table
CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
  amount DECIMAL(10, 2) NOT NULL,
  currency VARCHAR(3) DEFAULT 'USD',
  payment_method VARCHAR(50), -- card, bank_transfer, crypto
  status VARCHAR(50) DEFAULT 'pending', -- pending, processing, successful, failed, refunded
  gateway_reference VARCHAR(255),
  risk_score INTEGER,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  processed_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_booking ON payments(booking_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_gateway_ref ON payments(gateway_reference);

-- Fraud alerts table
CREATE TABLE fraud_alerts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id UUID REFERENCES bookings(id) ON DELETE CASCADE,
  traveler_id UUID NOT NULL REFERENCES travelers(id) ON DELETE CASCADE,
  risk_category VARCHAR(50), -- device_mismatch, velocity, anomaly
  risk_score INTEGER,
  is_blocked BOOLEAN DEFAULT FALSE,
  advisor_override_id UUID REFERENCES advisors(id) ON DELETE SET NULL,
  override_reason TEXT,
  detection JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  resolved_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_fraud_alerts_traveler ON fraud_alerts(traveler_id);
CREATE INDEX idx_fraud_alerts_booking ON fraud_alerts(booking_id);
CREATE INDEX idx_fraud_alerts_risk_score ON fraud_alerts(risk_score);
CREATE INDEX idx_fraud_alerts_is_blocked ON fraud_alerts(is_blocked);

-- Supplier rates cache table
CREATE TABLE supplier_rates (
  id BIGSERIAL PRIMARY KEY,
  supplier_id VARCHAR(100) NOT NULL,
  product_key VARCHAR(255) NOT NULL,
  price DECIMAL(10, 2),
  availability INTEGER,
  currency VARCHAR(3) DEFAULT 'USD',
  expires_at TIMESTAMP WITH TIME ZONE,
  fetched_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_supplier_rates_product ON supplier_rates(supplier_id, product_key, expires_at);
CREATE INDEX idx_supplier_rates_expires ON supplier_rates(expires_at);

-- Audit log table
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  entity_type VARCHAR(50) NOT NULL,
  entity_id VARCHAR(100) NOT NULL,
  action VARCHAR(50) NOT NULL, -- create, update, delete
  user_id VARCHAR(100),
  changes JSONB DEFAULT '{}'::jsonb,
  ip_address INET,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);

-- Suppliers table
CREATE TABLE suppliers (
  id VARCHAR(100) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  supplier_type VARCHAR(50), -- airline, hotel, activity
  api_key VARCHAR(255) NOT NULL UNIQUE,
  commission_rate DECIMAL(5, 2),
  is_active BOOLEAN DEFAULT TRUE,
  metadata JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_suppliers_active ON suppliers(is_active);
CREATE INDEX idx_suppliers_type ON suppliers(supplier_type);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply update_updated_at trigger to tables
CREATE TRIGGER trigger_update_travelers_updated_at
BEFORE UPDATE ON travelers
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_update_advisors_updated_at
BEFORE UPDATE ON advisors
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_update_bookings_updated_at
BEFORE UPDATE ON bookings
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_update_payments_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

-- Comments for documentation
COMMENT ON TABLE travelers IS 'Stores customer/traveler information with risk scoring and KYC status';
COMMENT ON TABLE advisors IS 'Stores travel advisor information and performance metrics';
COMMENT ON TABLE bookings IS 'Stores complete booking records with itinerary details';
COMMENT ON TABLE booking_items IS 'Stores individual components of bookings (flights, hotels, etc.)';
COMMENT ON TABLE payments IS 'Stores payment transaction details with fraud scoring';
COMMENT ON TABLE fraud_alerts IS 'Stores fraud detection alerts and advisor overrides';
COMMENT ON TABLE supplier_rates IS 'Caches supplier rates for quick access';
COMMENT ON TABLE audit_logs IS 'Stores audit trail of all system actions';
COMMENT ON TABLE suppliers IS 'Stores external supplier information and credentials';
