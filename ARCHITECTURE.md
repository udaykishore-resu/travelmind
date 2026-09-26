# TravelMind Architecture Design Document

## 1. System Overview

### 1.1 Problem Statement
Traditional travel agencies face:
- **OTA Competition**: Online travel agencies (Booking.com, Expedia) dominate through scale and AI-driven pricing
- **Margin Erosion**: Commission structures decline as consolidation increases
- **Trust Deficit**: Rise of scams and illegitimate actors damages credibility
- **Technology Debt**: Legacy systems struggle with real-time data and personalization
- **Advisor Relevance**: Manual processes can't keep pace with AI, limiting human value proposition

### 1.2 Solution Approach
TravelMind positions human advisors as **strategic consultants**, not commodity clerks, by:
1. **Automating low-value work** (rate comparison, basic bookings) via AI
2. **Augmenting advisor expertise** (Claude-powered copilot) with real-time recommendations
3. **Building direct client relationships** (proprietary CRM) to bypass OTA margins
4. **Enabling complex itineraries** (multi-destination, high-risk) where humans excel
5. **Maximizing margins** through intelligent pricing and supplier negotiation

## 2. Service Architecture

### 2.1 Microservices Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                      Clients (Web, Mobile, Voice)                │
└────────────────────────┬────────────────────────────────────────┘
                         │
        ┌────────────────▼────────────────┐
        │    Kong API Gateway              │
        │  (Auth, Rate Limit, Router)      │
        └────────────────┬────────────────┘
                         │
        ┌────────────────┴─────────────────────┐
        │                                       │
┌───────▼──────────────┐        ┌──────────────▼────────────┐
│ Booking Orchestrator  │        │ Advisor Service           │
│ ────────────────      │        │ ────────────────          │
│ • Rate aggregation    │        │ • Advisor management      │
│ • Hold management     │        │ • CRM integration         │
│ • Saga coordination   │        │ • Availability routing    │
│ • Confirmation track  │        │ • Commission tracking     │
└───────┬──────────────┘        └──────────────┬────────────┘
        │                                       │
        │                   ┌───────────────────┼──────────────────┐
        │                   │                   │                  │
┌───────▼──────┐  ┌────────▼──────┐  ┌────────▼────────┐  ┌──────▼─────┐
│AI Agent       │  │Personalization│  │Fraud Detection  │  │Supplier Sync│
│Service        │  │Engine          │  │                 │  │             │
│──────────     │  │────────────────│  │─────────────    │  │────────────│
│•Claude API    │  │•Segmentation   │  │•Risk scoring    │  │•Rate pulls │
│•Itinerary     │  │•Recommendations│  │•Anomaly detect  │  │•Cache mgmt │
│ build         │  │•Personalization│  │•Policy engine   │  │•Fallback   │
│•Complex req.  │  │•A/B testing    │  │•Advisor override│  │logic       │
└───────┬──────┘  └────────┬──────┘  └────────┬────────┘  └──────┬─────┘
        │                   │                   │                  │
        └───────────────────┼───────────────────┼──────────────────┘
                            │                   │
                   ┌────────▼───────┐  ┌───────▼──────┐
                   │ Payment         │  │ Notification │
                   │ Orchestrator    │  │ Service      │
                   │ (PCI-DSS L1)    │  │ (SMS, Email) │
                   └────────┬───────┘  └───────┬──────┘
                            │                   │
                     ┌──────▼───────────────────▼──────┐
                     │ Event Stream (Pub/Sub)          │
                     │ • Booking events                │
                     │ • Payment confirmations         │
                     │ • Fraud alerts                  │
                     │ • Supplier rate changes         │
                     └──────┬────────────────────┬─────┘
                            │                    │
                   ┌────────▼──┐      ┌─────────▼────┐
                   │ Analytics  │      │ Audit Logs   │
                   │ Pipeline   │      │ & Webhooks   │
                   └────────────┘      └──────────────┘
```

### 2.2 Service Responsibilities

| Service | Port | Language | Database | Key Domain |
|---------|------|----------|----------|-----------|
| API Gateway | 8080 | Go | - | Request routing, auth |
| Booking Orchestrator | 8081 | Go | PostgreSQL | Booking lifecycle |
| Advisor Service | 8082 | Go | PostgreSQL | Advisor/Customer mgmt |
| AI Agent | 8083 | Go | Firestore | LLM interaction |
| Personalization | 8084 | Go | BigQuery | ML recommendations |
| Fraud Detection | 8085 | Go | PostgreSQL | Risk scoring |
| Supplier Sync | 8086 | Go | Redis | Rate synchronization |
| Payment | 8087 | Go | PostgreSQL | Payment processing |

## 3. Data Architecture

### 3.1 Database Selection

```
┌────────────────────────────────────────────────────────┐
│                    Data Layer                          │
├─────────────────┬─────────────────┬───────────────────┤
│   PostgreSQL    │   Firestore     │      Redis        │
│   (OLTP)        │   (Document)    │    (Cache/Queue)  │
├─────────────────┼─────────────────┼───────────────────┤
│ • Bookings      │ • Chat history  │ • Rate cache      │
│ • Payments      │ • Traveler prof │ • Session data    │
│ • Suppliers     │ • Preferences   │ • Job queues      │
│ • Advisors      │ • Itineraries   │ • Circuit breaker │
│ • Transactions  │                 │ • Locks           │
│ • Audit logs    │                 │                   │
│ • Inventory     │                 │                   │
└─────────────────┴─────────────────┴───────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
    ┌───▼──────┐  ┌────▼─────┐  ┌──────▼────┐
    │ BigQuery  │  │ Memcached│  │ Elasticsearch│
    │ Analytics │  │ L2 Cache │  │ Search Index│
    │ Warehouse │  │          │  │             │
    └───────────┘  └──────────┘  └─────────────┘
```

### 3.2 PostgreSQL Schema (Core Tables)

```sql
-- Travelers
CREATE TABLE travelers (
  id UUID PRIMARY KEY,
  email STRING NOT NULL UNIQUE,
  phone_country STRING,
  profile_segment VARCHAR(50), -- luxury, budget, adventure, family, business
  risk_score INT DEFAULT 0,
  kyc_verified BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP
);

-- Advisors
CREATE TABLE advisors (
  id UUID PRIMARY KEY,
  name STRING NOT NULL,
  email STRING NOT NULL UNIQUE,
  specialization VARCHAR(100), -- cruise, luxury, adventure, corporate
  availability_status VARCHAR(20), -- available, busy, offline
  commission_rate DECIMAL(5,2),
  rating DECIMAL(3,2) DEFAULT 5.0,
  created_at TIMESTAMP
);

-- Bookings
CREATE TABLE bookings (
  id UUID PRIMARY KEY,
  traveler_id UUID NOT NULL REFERENCES travelers,
  advisor_id UUID REFERENCES advisors,
  status VARCHAR(50), -- draft, confirmed, paid, cancelled, completed
  total_price DECIMAL(10,2),
  commission DECIMAL(10,2),
  booking_window_days INT, -- days from creation to travel_date
  travel_date DATE NOT NULL,
  created_at TIMESTAMP,
  confirmed_at TIMESTAMP,
  
  -- JSON for flexible itinerary
  itinerary JSONB,
  
  INDEX idx_traveler (traveler_id),
  INDEX idx_status_date (status, travel_date)
);

-- Booking Items (flights, hotels, activities)
CREATE TABLE booking_items (
  id UUID PRIMARY KEY,
  booking_id UUID NOT NULL REFERENCES bookings,
  supplier_id VARCHAR(100) NOT NULL, -- "amadeus_flight", "booking_hotel"
  supplier_reference VARCHAR(255), -- PNR, confirmation code
  item_type VARCHAR(50), -- flight, hotel, activity, transfer
  price DECIMAL(10,2),
  status VARCHAR(50), -- pending, confirmed, failed, cancelled
  supplier_response JSONB, -- Raw API response from supplier
  created_at TIMESTAMP
);

-- Supplier Rates Cache
CREATE TABLE supplier_rates (
  id BIGSERIAL PRIMARY KEY,
  supplier_id VARCHAR(100) NOT NULL,
  product_key VARCHAR(255), -- "NYC-LAX-2026-09-30"
  price DECIMAL(10,2),
  availability INT,
  currency VARCHAR(3) DEFAULT 'USD',
  expires_at TIMESTAMP,
  fetched_at TIMESTAMP,
  
  INDEX idx_product (supplier_id, product_key, expires_at)
);

-- Payments
CREATE TABLE payments (
  id UUID PRIMARY KEY,
  booking_id UUID NOT NULL REFERENCES bookings,
  amount DECIMAL(10,2),
  currency VARCHAR(3),
  payment_method VARCHAR(50), -- card, bank_transfer, crypto
  status VARCHAR(50), -- pending, processing, successful, failed, refunded
  gateway_reference VARCHAR(255), -- Stripe/Adyen transaction ID
  risk_score INT,
  created_at TIMESTAMP,
  processed_at TIMESTAMP
);

-- Fraud Alerts
CREATE TABLE fraud_alerts (
  id UUID PRIMARY KEY,
  booking_id UUID REFERENCES bookings,
  traveler_id UUID NOT NULL REFERENCES travelers,
  risk_category VARCHAR(50), -- device_mismatch, velocity, anomaly
  risk_score INT,
  is_blocked BOOLEAN DEFAULT FALSE,
  advisor_override_id UUID REFERENCES advisors,
  override_reason TEXT,
  created_at TIMESTAMP,
  resolved_at TIMESTAMP
);

-- Audit Log
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  entity_type VARCHAR(50),
  entity_id VARCHAR(100),
  action VARCHAR(50), -- create, update, delete
  user_id VARCHAR(100),
  changes JSONB,
  ip_address INET,
  created_at TIMESTAMP,
  
  INDEX idx_entity (entity_type, entity_id),
  INDEX idx_created (created_at DESC)
);
```

### 3.3 Firestore Collections (Flexible Data)

```
/travelers/{traveler_id}
├── profile: { email, name, preferences, ... }
├── conversations: [ { ai_agent_id, messages: [...], created_at } ]
├── saved_itineraries: [ { destination, travelers, budget, date } ]
└── search_history: [ { query, results_count, timestamp } ]

/advisors/{advisor_id}
├── availability: { calendar, busy_until, timezone }
├── client_list: [ { traveler_id, last_contacted, lifetime_value } ]
└── performance: { bookings_month, revenue_month, rating }

/bookings/{booking_id}
├── itinerary: { legs: [ { from, to, date, type } ], notes }
├── risk_assessment: { geopolitical, weather, health, overall_score }
└── communication: [ { from, message, timestamp, channel } ]
```

## 4. Key Workflows

### 4.1 Booking Creation (Saga Pattern)

```
Client Request
    │
    ├─► 1. Create Booking (DRAFT status)
    │      └─► Event: BookingCreated
    │
    ├─► 2. Validate Traveler
    │      ├─► KYC Check
    │      ├─► Risk Score Assessment
    │      └─► Event: TravelerValidated
    │
    ├─► 3. AI Agent Service (Async)
    │      ├─► Parse complex requirements
    │      ├─► Search suppliers in parallel
    │      ├─► Build multi-leg itinerary
    │      └─► Event: ItineraryBuilt
    │
    ├─► 4. Price & Hold
    │      ├─► Reserve inventory (30-min hold)
    │      ├─► Lock rates
    │      └─► Event: InventoryHeld
    │
    ├─► 5. Fraud Check
    │      ├─► ML Risk Scoring
    │      ├─► Policy Engine
    │      ├─► If blocked → Advisor Review
    │      └─► Event: FraudCheckPassed
    │
    ├─► 6. Payment Processing
    │      ├─► 3D Secure Challenge
    │      ├─► Gateway Processing
    │      ├─► Retry Logic (3x)
    │      └─► Event: PaymentSuccessful
    │
    ├─► 7. Supplier Confirmation
    │      ├─► Create PNRs with suppliers
    │      ├─► Collect confirmation numbers
    │      └─► Event: SupplierConfirmed
    │
    └─► Status: CONFIRMED
            └─► Send confirmation email/SMS
```

### 4.2 AI-Powered Booking Assistant

```
User: "I need a 10-day trip for 2 families (8 people total) in Bali, 
        Dec 15-25, budget ~$15k, need family-friendly resorts, 
        want to avoid crowds, both families have kids with dietary restrictions"

AI Agent Flow:
  1. Parse Requirements
     ├─ Travelers: 8 (2 families, 4 adults, 4 kids aged 5-12)
     ├─ Destination: Bali (filter luxury, family-friendly)
     ├─ Dates: Dec 15-25 (high season, adjust recommendations)
     ├─ Budget: $1875/person
     └─ Constraints: Dietary (gluten-free, nut allergies)

  2. Search in Parallel
     ├─ Flights: NYC-Denpasar (direct + 1-stop options)
     ├─ Hotels: Ubud family resorts, Seminyak luxury (20 options)
     ├─ Activities: Family-friendly (cooking class, monkey forest, water sports)
     ├─ Transfers: Private car + traditional jukung boat
     └─ Special Services: Dietary meal arrangements, childcare

  3. Risk Assessment
     ├─ Geopolitical: Low (Bali stable)
     ├─ Health: Check vaccinations, dengue advisory
     ├─ Weather: Rainy season, monsoon planning
     └─ Insurance: Quote trip insurance (8 travelers)

  4. Build Itinerary
     ├─ Day 1-3: Ubud (cultural, active rest, acclimatization)
     ├─ Day 4-7: Seminyak (beach, family resort amenities)
     ├─ Day 8-10: Canggu (water sports, adventure activities)
     └─ Contingency: Backup activities (weather-proof)

  5. Negotiate & Optimize
     ├─ Group rate for hotel (8 rooms, 3+ days = 15% discount)
     ├─ Package flight + hotel (save 5%)
     ├─ Commission optimization (high-margin activities)
     └─ Loyalty points allocation

  6. Present to Advisor
     ├─ 3 itinerary options (conservative, moderate, adventurous)
     ├─ Total price breakdown by traveler
     ├─ Risk/insurance recommendations
     └─ Optional upgrades (premium airlines, Michelin restaurants)
```

### 4.3 Fraud Detection Flow

```
Booking Received
    │
    ├─► Device Fingerprinting
    │    ├─ Browser/OS/Device ID
    │    ├─ IP Geolocation
    │    └─ Check against traveler history (30-day lookback)
    │
    ├─► Behavioral Analysis (Real-time)
    │    ├─ Booking velocity (# bookings per hour)
    │    ├─ Price sensitivity deviation
    │    ├─ Destination pattern anomaly
    │    └─ Time-of-day pattern (ML model)
    │
    ├─► Velocity Checks
    │    ├─ Cards: <3 distinct cards/traveler/day
    │    ├─ Bookings: <5 bookings/traveler/hour
    │    └─ Amount: <$50k/traveler/day
    │
    ├─► ML Anomaly Detection
    │    ├─ Isolation Forest (multivariate outliers)
    │    ├─ LSTM (temporal sequence patterns)
    │    └─ Gradient Boosting (risk score ensemble)
    │
    ├─► Policy-Based Rules
    │    ├─ IF price > $100k AND new_device → BLOCK
    │    ├─ IF velocity_score > 0.8 → CHALLENGE_3DS
    │    ├─ IF multiple_high_risk_signals → MANUAL_REVIEW
    │    └─ If all checks pass → APPROVE
    │
    ├─► Advisor Involvement (if blocked)
    │    ├─ Alert in advisor portal
    │    ├─ Call traveler for verification
    │    ├─ Override with evidence (voice recognition, docs)
    │    └─ Audit log entry (compliance)
    │
    └─► Decision: APPROVE / CHALLENGE / BLOCK
```

## 5. Deployment Architecture

### 5.1 Kubernetes Cluster Setup (GKE)

```
Region: us-central1 (Primary)
├── Production Cluster (n1-standard-4 nodes, 5-10 replicas)
│   ├── Namespace: travelmind (core services)
│   ├── Namespace: monitoring (Prometheus, Grafana)
│   ├── Namespace: istio-system (service mesh, optional)
│   └── Storage: GKE Persistent Volumes (PostgreSQL, Redis)
│
├── Staging Cluster (smaller, cost-optimized)
│   └── Canary deployments
│
└── Dev Cluster (auto-scaling off, low cost)
    └── Feature branch testing
```

### 5.2 High-Availability Architecture

```
                    ┌─────────────────────┐
                    │  Cloud Load Balancer│
                    │  (Global, HTTPS)    │
                    └──────────┬──────────┘
                               │
            ┌──────────────────┼──────────────────┐
            │                  │                  │
    ┌───────▼──────┐  ┌────────▼───────┐ ┌──────▼────────┐
    │ us-central1  │  │  us-east4 (DR) │ │ eu-west1 (EU) │
    │   Primary    │  │                 │ │                │
    └───────┬──────┘  └────────┬───────┘ └──────┬────────┘
            │                  │                │
      ┌─────▼────────┐    ┌────▼───────┐  ┌───▼──────┐
      │ API Gateway  │    │   Passive  │  │ Regional │
      │ (3 replicas) │    │  Failover  │  │ Replica  │
      └─────┬────────┘    └────┬───────┘  └───┬──────┘
            │                  │               │
      ┌─────▼────────────┐    ┌────────────────▼──┐
      │ PostgreSQL       │    │ Postgres Replication
      │ Primary (HA)     │    │ (Synchronous)
      │ (3-node cluster) │    │ Backup
      └──────────────────┘    └───────────────────┘
            │
      ┌─────▼──────────────┐
      │ Cloud SQL Backups  │
      │ (hourly, 30-day)   │
      └────────────────────┘
```

### 5.3 CI/CD Pipeline

```
Git Push (main branch)
    │
    ├─► GitHub Actions Triggered
    │    ├─ Unit Tests (Go, 80%+ coverage)
    │    ├─ Integration Tests (Docker Compose)
    │    ├─ Security Scan (Snyk, SAST)
    │    ├─ Build Docker Images
    │    ├─ Push to Google Container Registry
    │    └─ Trigger Cloud Build
    │
    ├─► Staging Deployment
    │    ├─ Blue-Green deployment
    │    ├─ Smoke tests
    │    ├─ Manual QA sign-off
    │    └─ Performance benchmarks
    │
    └─► Production Deployment
         ├─ Canary (20% traffic for 1 hour)
         │  ├─ Monitor error rate, latency
         │  ├─ If metrics OK → Promote to 100%
         │  └─ Else → Rollback
         │
         └─ Progressive rollout
            └─ 2 hours total to 100% traffic
```

## 6. Security Architecture

### 6.1 Defense-in-Depth

```
Layer 1: Perimeter
├─ Cloud Armor (DDoS protection)
├─ WAF rules (SQL injection, XSS)
└─ Rate limiting (1000 req/min per IP)

Layer 2: API Gateway
├─ OAuth2/OIDC authentication
├─ JWT with 1-hour expiry + refresh tokens
├─ RBAC (roles: admin, advisor, traveler)
└─ API key versioning + rotation

Layer 3: Service-to-Service
├─ mTLS (mutual TLS, Istio)
├─ Service accounts (Workload Identity)
├─ Signed requests (HMAC-SHA256)
└─ Network policies (egress restrictions)

Layer 4: Data Security
├─ Encryption at rest (AES-256, CloudKMS)
├─ Encryption in transit (TLS 1.3)
├─ Database encryption (transparent)
├─ Column-level encryption (PII: SSN, cards)
└─ Field-level masking in logs

Layer 5: Application
├─ SQL injection prevention (parameterized queries)
├─ CSRF tokens
├─ Secure headers (CSP, HSTS, X-Frame-Options)
└─ PII data retention policies
```

### 6.2 Compliance

- **PCI-DSS Level 1**: Card data tokenization, network isolation
- **GDPR**: Data minimization, right to be forgotten, consent management
- **SOC2 Type II**: Annual audit, access controls, change management
- **CCPA**: Consumer privacy rights, opt-out mechanisms

## 7. Performance Optimization

### 7.1 Caching Strategy

```
Request Flow:
Client
  │
  ├─► Check Redis L1 Cache (15-min TTL)
  │    ├─ Rates: cached by product_key
  │    ├─ Traveler prefs: cached by ID
  │    └─ Supplier availability: cached TTL = API freshness
  │
  ├─► Check PostgreSQL Materialized Views (hourly refresh)
  │    ├─ Popular destinations aggregate
  │    ├─ Advisor availability matrix
  │    └─ Pricing history for analysis
  │
  ├─► Direct Query (if cache miss)
  │    ├─ Indexed lookups (B-tree)
  │    └─ Connection pooling (PgBouncer: 500 connections)
  │
  └─► Response sent (with cache headers for HTTP)
       └─ Cache-Control: max-age=300, public
```

### 7.2 Database Optimization

```
SELECT strategies:
├─ Denormalization for read-heavy queries (travelers, advisor profiles)
├─ Materialized views for aggregates (booking counts, revenue)
├─ Partial indexes (status = 'pending' on large tables)
├─ Connection pooling (PgBouncer: 500 connections)
├─ Read replicas for analytics (async replication)
└─ Vacuum & ANALYZE (nightly)

Write optimization:
├─ Batch inserts (100 rows / batch)
├─ Async event processing (event log batching)
├─ Write-ahead logging (WAL) for durability
└─ Prepared statements (connection reuse)
```

### 7.3 Frontend Optimization

- **Code Splitting**: Route-based, reduce main bundle to <100KB
- **Lazy Loading**: Images, modals, heavy components
- **Service Worker**: Offline capability, cache-first for static assets
- **CDN**: CloudFront for asset delivery
- **Image Optimization**: WebP, AVIF, responsive sizes

## 8. Monitoring & Observability

### 8.1 Metrics

```
Business Metrics:
├─ Bookings created (daily, by advisor, by destination)
├─ Revenue (total, by source, by supplier)
├─ Conversion rate (search → booking)
├─ Average order value (AOV)
├─ Customer lifetime value (CLV)
└─ Repeat booking rate

Technical Metrics:
├─ API latency (p50, p95, p99)
├─ Error rate by service
├─ Database query latency
├─ Cache hit rate
├─ Supplier sync success rate
└─ Payment failure rate

Operational Metrics:
├─ Pod restart count
├─ Memory/CPU utilization
├─ Disk usage
├─ Network I/O
└─ PagerDuty incident count
```

### 8.2 Alerting Thresholds

```
CRITICAL:
├─ API latency p99 > 1s (immediate page)
├─ Error rate > 5% (page)
├─ Database down (page)
└─ Payment gateway unreachable (page)

WARNING:
├─ Latency p99 > 500ms (escalate in 10min)
├─ Error rate > 1% (escalate in 5min)
├─ Memory > 80% (check in 2 hours)
└─ Cache hit rate < 50%

INFO:
├─ Deployment started
├─ Feature flag toggled
└─ Supplier sync completed
```

## 9. Scalability & Future Growth

### 9.1 Scaling Scenarios

```
Scenario A: 10x Traffic Growth (100k → 1M bookings/month)
├─ Horizontal Pod Autoscaling (HPA) → API Gateway 20+ replicas
├─ Database: PostgreSQL connection pool tuning
├─ Redis: Cluster mode (16 shards)
├─ Kafka: 10 partitions per topic
└─ Cost: ~$50k/month → $150k/month

Scenario B: New Supplier Integration (10 → 100 suppliers)
├─ Supplier sync: Async worker pools (50+ workers)
├─ Rate cache explosion: TTL reduction, eviction policy
├─ Inventory conflicts: Distributed locks (Redis)
└─ Monitoring: New supplier dashboards

Scenario C: International Expansion (US → 20 countries)
├─ Multi-currency: Forex rates API (daily refresh)
├─ Regional deployments: Edge CDN, regional databases
├─ Compliance: GDPR, local PCI versions, data residency
├─ Localization: 15+ languages (i18n framework)
└─ Cost: Multiplied infrastructure
```

### 9.2 Tech Debt Management

- **Code Reviews**: 2+ reviewers, automated checks (golangci-lint)
- **Test Coverage**: 80%+ threshold, enforced in CI
- **Dependencies**: Dependabot for vulnerability updates
- **Documentation**: RFCs for major decisions, ADRs for tech choices
- **Refactoring**: 20% of sprint capacity reserved

## 10. Disaster Recovery

### 10.1 RTO & RPO Targets

| Component | RTO | RPO |
|-----------|-----|-----|
| API Services | 15 min | 5 min |
| PostgreSQL | 30 min | 1 min |
| Customer Data | 1 hour | 5 min |
| Audit Logs | 4 hours | 15 min |

### 10.2 Failover Procedures

```
Primary Region Down (us-central1)
├─ Automated: Health check detects failure (2 min)
├─ DNS switchover to DR region (5 min propagation)
├─ Promote read replica to primary (3 min)
├─ Sync cached data from backups (5 min)
├─ Manual: Verify data integrity
└─ Post-incident: RCA within 24 hours
```

---

**Version**: 1.0
**Last Updated**: 2026-09-26
**Owner**: Uday Resu (@udaykishoreresu)
