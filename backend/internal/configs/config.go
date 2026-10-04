package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	AppName     string
	AppPort     string
	DatabaseURL string
	AppURL      string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	JWTAccessSecret     string
	JWTAccessTTLMinutes int
	JWTRefreshTTLDays   int

	PasswordResetTTLMinutes int
	RedisURL                string

	CORSAllowedOrigins   []string
	CORSAllowCredentials bool

	RateLimitGlobalRequests      int
	RateLimitGlobalWindowSeconds int
	RateLimitAuthRequests        int
	RateLimitAuthWindowSeconds   int
}

func Load() (*Config, error) {

	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))

	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	corsAllowedOriginsRaw := strings.TrimSpace(
		os.Getenv("CORS_ALLOWED_ORIGINS"),
	)

	if corsAllowedOriginsRaw == "" {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS is required")
	}

	corsAllowedOrigins := make([]string, 0)

	for _, origin := range strings.Split(
		corsAllowedOriginsRaw,
		",",
	) {
		origin = strings.TrimSpace(origin)

		if origin == "" {
			continue
		}

		corsAllowedOrigins = append(
			corsAllowedOrigins,
			origin,
		)
	}

	if len(corsAllowedOrigins) == 0 {
		return nil, fmt.Errorf(
			"at least one CORS allowed origin is required",
		)
	}

	corsAllowCredentialsRaw := getEnv(
		"CORS_ALLOW_CREDENTIALS",
		"false",
	)

	corsAllowCredentials, err := strconv.ParseBool(
		corsAllowCredentialsRaw,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"invalid CORS_ALLOW_CREDENTIALS",
		)
	}

	globalRequestsRaw := getEnv(
		"RATE_LIMIT_GLOBAL_REQUESTS",
		"120",
	)

	globalRequests, err := strconv.Atoi(
		globalRequestsRaw,
	)

	if err != nil || globalRequests <= 0 {
		return nil, fmt.Errorf(
			"invalid RATE_LIMIT_GLOBAL_REQUESTS",
		)
	}

	globalWindowRaw := getEnv(
		"RATE_LIMIT_GLOBAL_WINDOW_SECONDS",
		"60",
	)

	globalWindow, err := strconv.Atoi(
		globalWindowRaw,
	)

	if err != nil || globalWindow <= 0 {
		return nil, fmt.Errorf(
			"invalid RATE_LIMIT_GLOBAL_WINDOW_SECONDS",
		)
	}

	authRequestsRaw := getEnv(
		"RATE_LIMIT_AUTH_REQUESTS",
		"10",
	)

	authRequests, err := strconv.Atoi(
		authRequestsRaw,
	)

	if err != nil || authRequests <= 0 {
		return nil, fmt.Errorf(
			"invalid RATE_LIMIT_AUTH_REQUESTS",
		)
	}

	authWindowRaw := getEnv(
		"RATE_LIMIT_AUTH_WINDOW_SECONDS",
		"60",
	)

	authWindow, err := strconv.Atoi(
		authWindowRaw,
	)

	if err != nil || authWindow <= 0 {
		return nil, fmt.Errorf(
			"invalid RATE_LIMIT_AUTH_WINDOW_SECONDS",
		)
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	passwordResetTTLMinutesString := getEnv(
		"PASSWORD_RESET_TTL_MINUTES",
		"30",
	)

	passwordResetTTLMinutes, err := strconv.Atoi(
		passwordResetTTLMinutesString,
	)

	if err != nil || passwordResetTTLMinutes <= 0 {
		return nil, fmt.Errorf(
			"invalid PASSWORD_RESET_TTL_MINUTES",
		)
	}
	if smtpHost == "" {
		return nil, fmt.Errorf("SMTP_HOST is required")
	}

	if smtpPort == "" {
		return nil, fmt.Errorf("SMTP_PORT is required")
	}

	if smtpUsername == "" {
		return nil, fmt.Errorf("SMTP_USERNAME is required")
	}

	if smtpPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD is required")
	}

	if smtpFrom == "" {
		return nil, fmt.Errorf("SMTP_FROM is required")
	}

	appURL := os.Getenv("APP_URL")

	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")

	if jwtAccessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}

	jwtAccessTTLMinutesString := getEnv(
		"JWT_ACCESS_TTL_MINUTES",
		"15",
	)

	jwtAccessTTLMinutes, err := strconv.Atoi(
		jwtAccessTTLMinutesString,
	)

	if err != nil || jwtAccessTTLMinutes <= 0 {
		return nil, fmt.Errorf(
			"invalid JWT_ACCESS_TTL_MINUTES",
		)
	}

	jwtRefreshTTLDaysString := getEnv(
		"JWT_REFRESH_TTL_DAYS",
		"30",
	)

	jwtRefreshTTLDays, err := strconv.Atoi(
		jwtRefreshTTLDaysString,
	)

	if err != nil || jwtRefreshTTLDays <= 0 {
		return nil, fmt.Errorf(
			"invalid JWT_REFRESH_TTL_DAYS",
		)
	}

	return &Config{
		AppEnv:      getEnv("APP_ENV", "devlopment"),
		AppName:     getEnv("APP_NAME", "forensic_api"),
		AppPort:     port,
		DatabaseURL: databaseURL,

		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: smtpUsername,
		SMTPPassword: smtpPassword,
		SMTPFrom:     smtpFrom,
		AppURL:       appURL,

		JWTAccessSecret:         jwtAccessSecret,
		JWTAccessTTLMinutes:     jwtAccessTTLMinutes,
		JWTRefreshTTLDays:       jwtRefreshTTLDays,
		PasswordResetTTLMinutes: passwordResetTTLMinutes,
		RedisURL:                redisURL,

		CORSAllowedOrigins:   corsAllowedOrigins,
		CORSAllowCredentials: corsAllowCredentials,

		RateLimitGlobalRequests:      globalRequests,
		RateLimitGlobalWindowSeconds: globalWindow,
		RateLimitAuthRequests:        authRequests,
		RateLimitAuthWindowSeconds:   authWindow,
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
