# Multi-stage build for TravelMind API Gateway

# Stage 1: Builder
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" \
    -o /app/api-gateway ./cmd/api-gateway/main.go

# Stage 2: Runtime
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Copy binary from builder
COPY --from=builder /app/api-gateway /app/api-gateway

# Create non-root user for security
RUN addgroup -g 1000 travelmind && adduser -D -u 1000 -G travelmind travelmind
USER travelmind

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["/app/api-gateway"]
