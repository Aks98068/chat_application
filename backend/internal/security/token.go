package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ------------------------------------------------------------
// Secure Random Token
// ------------------------------------------------------------

func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be greater than zero")
	}

	token := make([]byte, length)

	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf(
			"failed to generate secure token: %w",
			err,
		)
	}

	return hex.EncodeToString(token), nil
}

// ------------------------------------------------------------
// Token Hashing
// ------------------------------------------------------------

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}

// ------------------------------------------------------------
// JWT Access Token
// ------------------------------------------------------------

type AccessTokenClaims struct {
	UserID    uuid.UUID       `json:"user_id"`
	Username  string          `json:"username"`
	Role      models.UserRole `json:"role"`
	TokenType string          `json:"token_type"`

	jwt.RegisteredClaims
}

// ------------------------------------------------------------
// Token Service
// ------------------------------------------------------------

type TokenService struct {
	accessSecret []byte
	accessTTL    time.Duration
	refreshTTL   time.Duration
}

// ------------------------------------------------------------
// Constructor
// ------------------------------------------------------------

func NewTokenService(
	accessSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *TokenService {

	return &TokenService{
		accessSecret: []byte(accessSecret),
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
	}
}

// ------------------------------------------------------------
// Access Token TTL
// ------------------------------------------------------------

func (s *TokenService) AccessTokenTTL() time.Duration {
	return s.accessTTL
}

// ------------------------------------------------------------
// Generate Access Token
// ------------------------------------------------------------

func (s *TokenService) GenerateAccessToken(
	user *models.User,
) (string, error) {

	now := time.Now()

	expiresAt := now.Add(s.accessTTL)

	claims := AccessTokenClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: "access",

		RegisteredClaims: jwt.RegisteredClaims{
			ID: uuid.New().String(),

			Subject: user.ID.String(),

			Issuer: "forensics-api",

			Audience: []string{
				"forensics-client",
			},

			IssuedAt: jwt.NewNumericDate(now),

			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(
		s.accessSecret,
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to sign access token: %w",
			err,
		)
	}

	return signedToken, nil
}

// ------------------------------------------------------------
// Generate Refresh Token
// ------------------------------------------------------------

func (s *TokenService) GenerateRefreshToken() (
	string,
	string,
	time.Time,
	error,
) {

	// 32 random bytes = 256 bits of entropy.
	rawToken, err := GenerateSecureToken(32)

	if err != nil {
		return "", "", time.Time{}, fmt.Errorf(
			"failed to generate refresh token: %w",
			err,
		)
	}

	// Only the hash is stored in the database.
	tokenHash := HashToken(rawToken)

	expiresAt := time.Now().Add(
		s.refreshTTL,
	)

	return rawToken, tokenHash, expiresAt, nil
}
