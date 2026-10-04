package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/Aks98068/forensics/internal/repository"
	"github.com/Aks98068/forensics/internal/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthTokenService struct {
	tokenService           *security.TokenService
	refreshTokenRepository *repository.RefreshTokenRepository
}

func NewAuthTokenService(
	tokenService *security.TokenService,
	refreshTokenRepository *repository.RefreshTokenRepository,
) *AuthTokenService {
	return &AuthTokenService{
		tokenService:           tokenService,
		refreshTokenRepository: refreshTokenRepository,
	}
}

type AuthenticationTokens struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresIn  int64
	RefreshTokenExpiresIn int64
}

func (s *AuthTokenService) CreateLoginTokens(
	ctx context.Context,
	user *models.User,
	userAgent string,
	ipAddress string,
) (*AuthenticationTokens, error) {

	// Generate the short-lived access token.
	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to generate access token: %w",
			err,
		)
	}

	// Generate the long-lived refresh token.
	rawRefreshToken, tokenHash, refreshExpiresAt, err :=
		s.tokenService.GenerateRefreshToken()

	if err != nil {
		return nil, fmt.Errorf(
			"failed to generate refresh token: %w",
			err,
		)
	}

	// Every login starts a new refresh-token family.
	tokenFamily := uuid.New()

	refreshToken := &models.RefreshToken{
		UserID:      user.ID,
		TokenHash:   tokenHash,
		TokenFamily: tokenFamily,
		ExpiresAt:   refreshExpiresAt,
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
	}

	// This is a new database record, so use Create.
	if err := s.refreshTokenRepository.Create(
		ctx,
		refreshToken,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to store refresh token: %w",
			err,
		)
	}

	return &AuthenticationTokens{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,

		AccessTokenExpiresIn: int64(
			s.tokenService.AccessTokenTTL().Seconds(),
		),

		RefreshTokenExpiresIn: int64(
			time.Until(refreshExpiresAt).Seconds(),
		),
	}, nil
}

func (s *AuthTokenService) Refresh(
	ctx context.Context,
	rawRefreshToken string,
	userAgent string,
	ipAddress string,
) (*AuthenticationTokens, error) {

	// The client must provide a refresh token.
	if rawRefreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	// Never search the database using the raw token.
	// Only its SHA-256 hash is stored in the database.
	tokenHash := security.HashToken(rawRefreshToken)

	// Generate a completely new refresh token.
	//
	// The raw token is returned to the client.
	// Only the hash is stored in MySQL.
	rawNewRefreshToken, newTokenHash, newRefreshExpiresAt, err :=
		s.tokenService.GenerateRefreshToken()

	if err != nil {
		return nil, fmt.Errorf(
			"failed to generate new refresh token: %w",
			err,
		)
	}

	/*
		The repository will fill in:

			UserID
			TokenFamily
			ID

		from the old token after acquiring the
		database row lock.
	*/
	newRefreshToken := &models.RefreshToken{
		TokenHash: newTokenHash,
		ExpiresAt: newRefreshExpiresAt,
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	now := time.Now()

	/*
		Rotate performs the security-critical operation
		inside one database transaction.

		It:

		1. Locks the old refresh-token row.
		2. Checks expiration.
		3. Checks revocation.
		4. Detects token reuse.
		5. Creates the replacement token.
		6. Revokes the old token.
		7. Links old -> new.
		8. Commits the transaction.
	*/
	rotation, err := s.refreshTokenRepository.Rotate(
		ctx,
		tokenHash,
		newRefreshToken,
		now,
	)

	if err != nil {

		// The token does not exist.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefreshToken
		}

		return nil, fmt.Errorf(
			"failed to rotate refresh token: %w",
			err,
		)
	}

	// The refresh token has expired.
	if rotation.TokenExpired {
		return nil, ErrRefreshTokenExpired
	}

	// The refresh token was already revoked.
	if rotation.TokenRevoked {
		return nil, ErrRefreshTokenRevoked
	}

	/*
		The token was already rotated before.

		Someone is attempting to reuse an old
		refresh token.

		The repository will have revoked the
		remaining active tokens in that family.
	*/
	if rotation.ReuseDetected {
		return nil, ErrRefreshTokenReuse
	}

	/*
		These should never be nil after a successful
		rotation.

		Keep the check anyway so the service never
		dereferences a nil pointer.
	*/
	if rotation.OldToken == nil ||
		rotation.NewToken == nil {

		return nil, ErrInvalidRefreshToken
	}

	/*
		Generate a new access token for the user
		associated with the refresh token.
	*/
	accessToken, err :=
		s.tokenService.GenerateAccessToken(
			&rotation.OldToken.User,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to generate new access token: %w",
			err,
		)
	}

	return &AuthenticationTokens{
		AccessToken:  accessToken,
		RefreshToken: rawNewRefreshToken,

		AccessTokenExpiresIn: int64(
			s.tokenService.AccessTokenTTL().Seconds(),
		),

		RefreshTokenExpiresIn: int64(
			time.Until(newRefreshExpiresAt).Seconds(),
		),
	}, nil
}

// ============================================================
// Logout
// ============================================================
//
// Revokes the refresh token associated with the current
// authentication session.
//
// Logout is intentionally idempotent.
//
// ============================================================

func (s *AuthTokenService) Logout(
	ctx context.Context,
	rawRefreshToken string,
) error {

	// --------------------------------------------------------
	// Validate input.
	// --------------------------------------------------------

	rawRefreshToken = strings.TrimSpace(
		rawRefreshToken,
	)

	if rawRefreshToken == "" {
		return nil
	}

	// --------------------------------------------------------
	// Prevent unnecessarily large input.
	//
	// GenerateSecureToken(32) produces 64 hexadecimal
	// characters, so anything dramatically larger than that
	// is not a legitimate token generated by our server.
	// --------------------------------------------------------

	const maximumRefreshTokenLength = 512

	if len(rawRefreshToken) > maximumRefreshTokenLength {
		return nil
	}

	// --------------------------------------------------------
	// Never store the raw refresh token in the database.
	//
	// Hash it first.
	// --------------------------------------------------------

	tokenHash := security.HashToken(
		rawRefreshToken,
	)

	// --------------------------------------------------------
	// Revoke the token.
	// --------------------------------------------------------

	err := s.refreshTokenRepository.RevokeByTokenHash(
		ctx,
		tokenHash,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf(
			"failed to revoke refresh token: %w",
			err,
		)
	}

	return nil
}
