package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const contentSecurityPolicy = "default-src 'none'; " +
	"style-src 'self' https://cdn.jsdelivr.net; " +
	"script-src 'self' https://cdn.jsdelivr.net; " +
	"font-src 'self' https://cdn.jsdelivr.net; " +
	"img-src 'self' data:; " +
	"connect-src 'self'; " +
	"form-action 'self'; " +
	"base-uri 'self'; " +
	"frame-ancestors 'none'"

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header(
			"X-Content-Type-Options",
			"nosniff",
		)

		c.Header(
			"X-Frame-Options",
			"DENY",
		)

		c.Header(
			"Referrer-Policy",
			"strict-origin-when-cross-origin",
		)

		c.Header(
			"Permissions-Policy",
			"camera=(), microphone=(), geolocation=(), payment=()",
		)

		c.Header(
			"Content-Security-Policy",
			contentSecurityPolicy,
		)

		c.Header(
			"Cross-Origin-Opener-Policy",
			"same-origin",
		)

		c.Header(
			"Cross-Origin-Resource-Policy",
			"same-origin",
		)

		c.Header(
			"X-Permitted-Cross-Domain-Policies",
			"none",
		)

		// no-store only for API responses, so CSS/JS can be cached.
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header(
				"Cache-Control",
				"no-store",
			)
		}

		c.Next()
	}
}
