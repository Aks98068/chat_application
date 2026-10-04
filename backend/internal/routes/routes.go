package routes

import (
	"github.com/Aks98068/forensics/internal/handlres"
	middleware "github.com/Aks98068/forensics/internal/middlewares"
	"github.com/Aks98068/forensics/internal/models"
	"github.com/gin-gonic/gin"
)

type Config struct {
	JWTAccessSecret string
	JWTIssuer       string
	JWTAudience     string
}

func Routes(
	router *gin.Engine,
	authHandler *handlres.AuthHandler,
	userHandler *handlres.UserHandler,
	cfg *Config,
) {

	// ============================================================
	// HEALTH CHECK
	// ============================================================

	router.GET(
		"/health",
		func(c *gin.Context) {
			c.JSON(
				200,
				gin.H{
					"status": "ok",
				},
			)
		},
	)

	// ============================================================
	// API V1
	// ============================================================

	api := router.Group(
		"/api/v1",
	)

	// ============================================================
	// AUTH ROUTES
	// ============================================================

	auth := api.Group(
		"/auth",
	)

	auth.POST(
		"/register",
		authHandler.Register,
	)

	auth.POST(
		"/verify-email",
		authHandler.VerifyEmail,
	)

	auth.POST(
		"/resend-verification",
		authHandler.ResendVerificationEmail,
	)

	auth.POST(
		"/login",
		authHandler.Login,
	)

	auth.POST(
		"/refresh",
		authHandler.Refresh,
	)

	auth.POST(
		"/logout",
		authHandler.Logout,
	)

	auth.POST(
		"/forgot-password",
		authHandler.ForgotPassword,
	)

	auth.POST(
		"/reset-password",
		authHandler.ResetPassword,
	)

	// ============================================================
	// PROTECTED ROUTES
	// ============================================================

	protected := api.Group(
		"/protected",
	)

	protected.Use(
		middleware.JWTAuthMiddleware(
			middleware.JWTAuthConfig{
				AccessSecret: cfg.JWTAccessSecret,
				Issuer:       cfg.JWTIssuer,
				Audience:     cfg.JWTAudience,
			},
		),
	)

	// ============================================================
	// USER ROUTES
	// ============================================================

	userRoutes := protected.Group(
		"/user",
	)

	userRoutes.Use(
		middleware.RequireRoles(
			models.RoleUser,
			models.RoleAnalyst,
			models.RoleAdmin,
		),
	)

	userRoutes.GET(
		"/me",
		userHandler.Me,
	)

	// ============================================================
	// ANALYST ROUTES
	// ============================================================

	analystRoutes := protected.Group(
		"/analyst",
	)

	analystRoutes.Use(
		middleware.RequireRoles(
			models.RoleAnalyst,
			models.RoleAdmin,
		),
	)

	// ============================================================
	// ADMIN ROUTES
	// ============================================================

	adminRoutes := protected.Group(
		"/admin",
	)

	adminRoutes.Use(
		middleware.RequireRoles(
			models.RoleAdmin,
		),
	)

	// Prevent unused-variable compiler errors until
	// actual analyst/admin endpoints are added.
	_ = analystRoutes
	_ = adminRoutes
}
