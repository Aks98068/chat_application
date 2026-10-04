package middlewares

import "github.com/gin-gonic/gin"

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
			"default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'",
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

		c.Header(
			"Cache-Control",
			"no-store",
		)

		c.Next()
	}
}
