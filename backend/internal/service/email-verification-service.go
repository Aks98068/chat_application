package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/Aks98068/forensics/internal/repository"
	"github.com/Aks98068/forensics/internal/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	verificationTokenLifetime  = 24 * time.Hour
	resendVerificationCooldown = 60 * time.Second
)

type EmailVerificationService struct {
	tokenRepository *repository.EmailVerificationRepository
	userRepository  *repository.UserRepository
	emailService    *EmailService
	appURL          string
}

func NewEmailVerificationService(
	tokenRepository *repository.EmailVerificationRepository,
	userRepository *repository.UserRepository,
	emailService *EmailService,
	appURL string,
) *EmailVerificationService {
	return &EmailVerificationService{
		tokenRepository: tokenRepository,
		userRepository:  userRepository,
		emailService:    emailService,
		appURL:          strings.TrimRight(appURL, "/"),
	}
}

// ------------------------------------------------------------
// Create verification token
// ------------------------------------------------------------

func (s *EmailVerificationService) CreateToken(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {

	rawToken, err := security.GenerateSecureToken(32)
	if err != nil {
		return "", err
	}

	tokenHash := security.HashToken(rawToken)

	token := &models.EmailVerificationToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(verificationTokenLifetime),
	}

	if err := s.tokenRepository.Create(ctx, token); err != nil {
		return "", err
	}

	return rawToken, nil
}

// ------------------------------------------------------------
// Verify email
// ------------------------------------------------------------

func (s *EmailVerificationService) VerifyEmail(
	ctx context.Context,
	rawToken string,
) error {

	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return ErrInvalidVerification
	}

	tokenHash := security.HashToken(rawToken)

	token, err := s.tokenRepository.FindByTokenHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidVerification
		}

		return err
	}

	if token.UsedAt != nil {
		return ErrVerificationUsed
	}

	if token.RevokedAt != nil {
		return ErrVerificationRevoked
	}

	if time.Now().After(token.ExpiresAt) {
		return ErrInvalidVerification
	}

	if err := s.userRepository.MarkEmailAsVerified(
		ctx,
		token.UserID,
	); err != nil {
		return err
	}

	now := time.Now()

	if err := s.tokenRepository.MarkAsUsed(
		ctx,
		token.ID,
		now,
	); err != nil {
		return err
	}

	return nil
}

// ------------------------------------------------------------
// Resend verification email
// ------------------------------------------------------------

func (s *EmailVerificationService) ResendVerificationEmail(
	ctx context.Context,
	email string,
) error {

	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" {
		return nil
	}

	user, err := s.userRepository.FindByEmail(
		ctx,
		email,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Do not reveal whether the email exists.
			return nil
		}

		return err
	}

	// Already verified.
	// Return nil so the API response remains generic.
	if user.EmailVerified {
		return nil
	}

	// --------------------------------------------------------
	// Check resend cooldown
	// --------------------------------------------------------

	latestToken, err := s.tokenRepository.FindLatestToken(
		ctx,
		user.ID,
	)

	if err != nil &&
		!errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if latestToken != nil {
		cooldownUntil := latestToken.CreatedAt.Add(
			resendVerificationCooldown,
		)

		if time.Now().Before(cooldownUntil) {
			return ErrResendTooSoon
		}
	}

	// --------------------------------------------------------
	// Generate new token
	// --------------------------------------------------------

	rawToken, err := security.GenerateSecureToken(32)
	if err != nil {
		return err
	}

	tokenHash := security.HashToken(rawToken)

	now := time.Now()

	// --------------------------------------------------------
	// Revoke previous active tokens
	// --------------------------------------------------------

	if err := s.tokenRepository.RevokeActiveTokens(
		ctx,
		user.ID,
		now,
	); err != nil {
		return err
	}

	// --------------------------------------------------------
	// Create new token
	// --------------------------------------------------------

	newToken := &models.EmailVerificationToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(verificationTokenLifetime),
	}

	if err := s.tokenRepository.Create(
		ctx,
		newToken,
	); err != nil {
		return err
	}

	// --------------------------------------------------------
	// Build verification URL
	// --------------------------------------------------------

	verificationURL := s.appURL +
		"/verify-email?token=" +
		rawToken

	// --------------------------------------------------------
	// Send email
	// --------------------------------------------------------

	if err := s.emailService.SendVerificationEmail(
		user.Email,
		verificationURL,
	); err != nil {
		return err
	}

	return nil
}
