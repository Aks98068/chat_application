package main

import (
	"log"
	"time"

	"github.com/Aks98068/forensics/internal/configs"
	"github.com/Aks98068/forensics/internal/database"
	"github.com/Aks98068/forensics/internal/frontend"
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
		log.Fatalf("configuration error: %v", err)
	}

	// ============================================================
	// DATABASE
	// ============================================================

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}

	log.Println("MySQL database connected successfully")

	if err := database.MigrateDB(db); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	log.Println("database migrated successfully")

	// ============================================================
	// REPOSITORIES
	// ============================================================

	userRepository := repository.NewUserRepository(db)

	emailVerificationRepository :=
		repository.NewEmailVerificationRepository(db)

	refreshTokenRepository :=
		repository.NewRefreshTokenRepository(db)

	passwordResetRepository :=
		repository.NewPasswordResetRepository(db)

	// ============================================================
	// SECURITY
	// ============================================================

	passwordHasher := security.Newpassword()

	tokenService := security.NewTokenService(
		cfg.JWTAccessSecret,
		time.Duration(cfg.JWTAccessTTLMinutes)*time.Minute,
		time.Duration(cfg.JWTRefreshTTLDays)*24*time.Hour,
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
	// EMAIL VERIFICATION
	// ============================================================

	emailVerificationService :=
		service.NewEmailVerificationService(
			emailVerificationRepository,
			userRepository,
			emailService,
			cfg.AppURL,
		)

	// ============================================================
	// PASSWORD RESET
	// ============================================================

	passwordResetService :=
		service.NewPasswordResetService(
			passwordResetRepository,
			userRepository,
			passwordHasher,
			emailService,
			cfg.AppURL,
			time.Duration(cfg.PasswordResetTTLMinutes)*time.Minute,
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
	// GIN
	// ============================================================

	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	// ============================================================
	// SECURITY MIDDLEWARE
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
	// TRUSTED PROXIES
	// ============================================================

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf(
			"trusted proxy configuration error: %v",
			err,
		)
	}

	// ============================================================
	// FRONTEND RENDERER
	// ============================================================

	/*
		IMPORTANT:

		The program is normally started from:

		C:\go-tools\chatt-application\backend

		Therefore:

		../frontend/public

		resolves to:

		C:\go-tools\chatt-application\frontend\public
	*/

	frontendPublicDir := "../frontend/public"

	frontendRenderer, err := frontend.NewRenderer(
		frontendPublicDir,
	)

	if err != nil {
		log.Fatalf(
			"frontend renderer initialization error: %v",
			err,
		)
	}

	frontendRenderer.SetAppName(
		"ChatApplication",
	)

	frontendRenderer.SetDescription(
		"Secure real-time chat application",
	)

	// In debug mode templates are re-parsed on every request,
	// so HTML edits show up on refresh without a restart.
	frontendRenderer.SetDevMode(gin.Mode() != gin.ReleaseMode)

	// Logs a warning for every page template that is missing.
	frontendRenderer.CheckPages(
		"home",
		"register",
		"login",
		"verify-email",
		"forgot-password",
		"reset-password",
		"404",
	)

	log.Printf(
		"frontend directory: %s",
		frontendRenderer.PublicDir(),
	)

	log.Println(
		"frontend renderer initialized successfully",
	)

	// ============================================================
	// ROUTES
	// ============================================================

	routes.Routes(
		router,
		authHandler,
		userHandler,
		frontendRenderer,
		&routes.Config{
			JWTAccessSecret: cfg.JWTAccessSecret,
			JWTIssuer:       "chat-api",
			JWTAudience:     "chat-client",
		},
	)

	// ============================================================
	// SERVER
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
