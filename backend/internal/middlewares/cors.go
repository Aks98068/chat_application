package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

func CORS(config CORSConfig) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{})

	for _, origin := range config.AllowedOrigins {
		origin = strings.TrimSpace(origin)

		if origin == "" {
			continue
		}

		allowedOrigins[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))

		// Only validate CORS when the browser sends an Origin header.
		if origin != "" {
			if _, allowed := allowedOrigins[origin]; !allowed {
				c.AbortWithStatusJSON(
					http.StatusForbidden,
					gin.H{
						"error": "origin not allowed",
					},
				)
				return
			}

			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")

			c.Header(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			)

			c.Header(
				"Access-Control-Allow-Headers",
				"Authorization, Content-Type, X-Request-ID",
			)

			c.Header(
				"Access-Control-Expose-Headers",
				"X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, Retry-After",
			)

			if config.AllowCredentials {
				c.Header(
					"Access-Control-Allow-Credentials",
					"true",
				)
			}

			// Tell the browser how long it can cache
			// the preflight response.
			c.Header(
				"Access-Control-Max-Age",
				"600",
			)
		}

		// Handle browser CORS preflight.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
