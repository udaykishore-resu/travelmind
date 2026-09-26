package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/revansystems/travelmind/internal/db"
	"github.com/revansystems/travelmind/internal/observability"
)

// HealthCheck returns health status of the service
func HealthCheck(pgDB *db.PostgresDB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := make(map[string]observability.CheckResult)

		// Check PostgreSQL
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := pgDB.Ping(ctx); err != nil {
			checks["postgres"] = observability.CheckResult{
				Status: "down",
				Error:  err.Error(),
			}
		} else {
			checks["postgres"] = observability.CheckResult{
				Status: "up",
			}
		}

		// Check Redis
		if err := redisClient.Ping(ctx).Err(); err != nil {
			checks["redis"] = observability.CheckResult{
				Status: "down",
				Error:  err.Error(),
			}
		} else {
			checks["redis"] = observability.CheckResult{
				Status: "up",
			}
		}

		// Determine overall status
		overallStatus := "up"
		for _, check := range checks {
			if check.Status == "down" {
				overallStatus = "degraded"
			}
		}

		result := observability.HealthCheckResult{
			Status:    overallStatus,
			Checks:    checks,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, result)
	}
}

// ReadinessCheck returns readiness status (for K8s)
func ReadinessCheck(pgDB *db.PostgresDB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		// Check critical dependencies
		if err := pgDB.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ready": false, "error": "Database not ready"})
			return
		}

		if err := redisClient.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ready": false, "error": "Redis not ready"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ready": true})
	}
}

// SwaggerUI returns swagger UI HTML
func SwaggerUI() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, swaggerHTML)
	}
}

var swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <title>TravelMind API Documentation</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@3/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/swagger.yaml",
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout"
      })
      window.ui = ui
    }
  </script>
</body>
</html>`
