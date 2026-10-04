package main

import (
	"log"
	"time"

	"github.com/Aks98068/forensics/internal/configs"
	"github.com/Aks98068/forensics/internal/database"
	"github.com/Aks98068/forensics/internal/handlres"
	middleware "github.com/Aks98068/forensics/internal/middlewares"
	"github.com/Aks98068/forensics/internal/repository"
	"github.com/Aks98068/forensics/internal/routes"
	"github.com/Aks98068/forensics/internal/security"
	"github.com/Aks98068/forensics/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {

	// ============================================================
	// CONFIGURATION
	// ============================================================

	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf(
			"configuration error: %v",
			err,
		)
	}

	// ============================================================
	// DATABASE
	// ============================================================

	db, err := database.Connect(
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatalf(
			"database error: %v",
			err,
		)
	}

	log.Println(
		"MySQL database connected successfully",
	)

	if err := database.MigrateDB(db); err != nil {
		log.Fatalf(
			"migration error: %v",
			err,
		)
	}

	log.Println(
		"database migrated successfully",
	)

	// ============================================================
	// REPOSITORIES
	// ============================================================

	userRepository := repository.NewUserRepository(
		db,
	)

	emailVerificationRepository :=
		repository.NewEmailVerificationRepository(
			db,
		)

	refreshTokenRepository :=
		repository.NewRefreshTokenRepository(
			db,
		)

	passwordResetRepository :=
		repository.NewPasswordResetRepository(
			db,
		)

	// ============================================================
	// SECURITY
	// ============================================================

	passwordHasher := security.Newpassword()

	tokenService := security.NewTokenService(
		cfg.JWTAccessSecret,
		time.Duration(
			cfg.JWTAccessTTLMinutes,
		)*time.Minute,
		time.Duration(
			cfg.JWTRefreshTTLDays,
		)*24*time.Hour,
	)

	// ============================================================
	// EMAIL SERVICE
	// ============================================================

	emailService := service.NewEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	// ============================================================
	// EMAIL VERIFICATION SERVICE
	// ============================================================

	emailVerificationService :=
		service.NewEmailVerificationService(
			emailVerificationRepository,
			userRepository,
			emailService,
			cfg.AppURL,
		)

	// ============================================================
	// PASSWORD RESET SERVICE
	// ============================================================

	passwordResetService :=
		service.NewPasswordResetService(
			passwordResetRepository,
			userRepository,
			passwordHasher,
			emailService,
			cfg.AppURL,
			time.Duration(
				cfg.PasswordResetTTLMinutes,
			)*time.Minute,
		)

	// ============================================================
	// AUTH TOKEN SERVICE
	// ============================================================

	authTokenService := service.NewAuthTokenService(
		tokenService,
		refreshTokenRepository,
	)

	// ============================================================
	// AUTH SERVICE
	// ============================================================

	authService := service.NewAuthService(
		userRepository,
		passwordHasher,
		emailVerificationService,
		emailService,
		authTokenService,
		cfg.AppURL,
		passwordResetService,
	)

	// ============================================================
	// USER SERVICE
	// ============================================================

	userService := service.NewUserService(
		userRepository,
	)

	// ============================================================
	// HANDLERS
	// ============================================================

	authHandler := handlres.NewAuthHandler(
		authService,
	)

	userHandler := handlres.NewUserHandler(
		userService,
	)

	// ============================================================
	// GIN ROUTER
	// ============================================================

	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	// ============================================================
	// GLOBAL SECURITY MIDDLEWARE
	// ============================================================

	router.Use(
		middleware.RequestID(),
	)

	router.Use(
		middleware.SecurityHeaders(),
	)

	router.Use(
		middleware.CORS(
			middleware.CORSConfig{
				AllowedOrigins:   cfg.CORSAllowedOrigins,
				AllowCredentials: cfg.CORSAllowCredentials,
			},
		),
	)

	router.Use(
		middleware.RequestBodyLimit(
			1024 * 1024,
		),
	)

	// ============================================================
	// TRUSTED PROXY CONFIGURATION
	// ============================================================

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf(
			"trusted proxy configuration error: %v",
			err,
		)
	}

	// ============================================================
	// ROUTES
	// ============================================================

	routes.Routes(
		router,
		authHandler,
		userHandler,
		&routes.Config{
			JWTAccessSecret: cfg.JWTAccessSecret,
			JWTIssuer:       "forensics-api",
			JWTAudience:     "forensics-client",
		},
	)

	// ============================================================
	// START SERVER
	// ============================================================

	address := ":" + cfg.AppPort

	log.Printf(
		"%s running on http://localhost%s",
		cfg.AppName,
		address,
	)

	if err := router.Run(address); err != nil {
		log.Fatalf(
			"server error: %v",
			err,
		)
	}
}
