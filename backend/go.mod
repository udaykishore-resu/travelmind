module github.com/udaykishore-resu/travelmind

go 1.21

require (
	// Web Framework
	github.com/gin-gonic/gin v1.9.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0

	// Database
	github.com/lib/pq v1.10.9
	github.com/jackc/pgx/v5 v5.5.0
	github.com/jmoiron/sqlc v1.25.0
	cloud.google.com/go/firestore v1.14.0
	github.com/redis/go-redis/v9 v9.4.0

	// AI Integration
	github.com/anthropics/sdk-go v0.1.0

	// Messaging & Events
	cloud.google.com/go/pubsub v1.33.0
	github.com/segmentio/kafka-go v0.4.46

	// ML / Analytics
	cloud.google.com/go/bigquery v1.56.0
	google.cloud.ai/go/aiplatform v1.58.0

	// Utilities
	github.com/google/uuid v1.5.0
	github.com/joho/godotenv v1.5.1
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.17.0

	// Logging & Observability
	github.com/sirupsen/logrus v1.9.3
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0
	github.com/prometheus/client_golang v1.17.0

	// Testing
	github.com/stretchr/testify v1.8.4
	github.com/testcontainers/testcontainers-go v0.27.0

	// HTTP Client
	github.com/resty/resty/v2 v2.10.0

	// Data Validation
	github.com/go-playground/validator/v10 v10.16.0

	// Crypto & Security
	golang.org/x/crypto v0.16.0
	github.com/golang-jwt/jwt/v5 v5.0.0

	// Google Cloud
	cloud.google.com/go v0.110.10
	google.golang.org/api v0.153.0

	// Misc
	github.com/fatih/color v1.16.0
	github.com/mitchellh/mapstructure v1.5.0
)

require (
	// Indirect dependencies
	github.com/golang/protobuf v1.5.3
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
)
