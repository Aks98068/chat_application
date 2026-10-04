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
	"gorm.io/gorm"
)

var (
	ErrPasswordResetInvalid = errors.New(
		"invalid or expired password reset token",
	)

	ErrPasswordResetTooSoon = errors.New(
		"password reset requested too soon",
	)
)

type PasswordResetService struct {
	passwordResetRepository *repository.PasswordResetRepository
	userRepository          *repository.UserRepository
	passwordHasher          *security.PasswordHash
	emailService            *EmailService
	appURL                  string
	tokenTTL                time.Duration
}

func NewPasswordResetService(
	passwordResetRepository *repository.PasswordResetRepository,
	userRepository *repository.UserRepository,
	passwordHasher *security.PasswordHash,
	emailService *EmailService,
	appURL string,
	tokenTTL time.Duration,
) *PasswordResetService {

	return &PasswordResetService{
		passwordResetRepository: passwordResetRepository,
		userRepository:          userRepository,
		passwordHasher:          passwordHasher,
		emailService:            emailService,
		appURL: strings.TrimRight(
			appURL,
			"/",
		),
		tokenTTL: tokenTTL,
	}
}

// ============================================================
// Request Password Reset
// ============================================================
//
// IMPORTANT:
//
// This method intentionally does not reveal whether the
// supplied identifier belongs to an account.
//
// The handler should return the same response regardless.
//
// ============================================================

func (s *PasswordResetService) RequestPasswordReset(
	ctx context.Context,
	identifier string,
) error {

	identifier = strings.ToLower(
		strings.TrimSpace(identifier),
	)

	if identifier == "" {
		return nil
	}

	// --------------------------------------------------------
	// Find account.
	// --------------------------------------------------------

	user, err := s.userRepository.FindByIdentifier(
		ctx,
		identifier,
	)

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {
			// Do not reveal account existence.
			return nil
		}

		return fmt.Errorf(
			"failed to find user: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Revoke previous active reset tokens.
	// --------------------------------------------------------

	now := time.Now()

	if err := s.passwordResetRepository.
		RevokeActiveByUserID(
			ctx,
			user.ID,
			now,
		); err != nil {

		return fmt.Errorf(
			"failed to revoke previous reset tokens: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Generate cryptographically secure reset token.
	// --------------------------------------------------------

	rawToken, err := security.GenerateSecureToken(
		32,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to generate password reset token: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Hash token before storing it.
	// --------------------------------------------------------

	tokenHash := security.HashToken(
		rawToken,
	)

	// --------------------------------------------------------
	// Create reset token.
	// --------------------------------------------------------

	expiresAt := now.Add(
		s.tokenTTL,
	)

	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	if err := s.passwordResetRepository.Create(
		ctx,
		resetToken,
	); err != nil {

		return fmt.Errorf(
			"failed to store password reset token: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Construct reset URL.
	// --------------------------------------------------------

	resetURL := fmt.Sprintf(
		"%s/reset-password?token=%s",
		s.appURL,
		rawToken,
	)

	// --------------------------------------------------------
	// Send email if the provided email service supports it.
	// Use a dynamic type assertion so this package does not
	// require a specific concrete EmailService method to exist.
	// --------------------------------------------------------

	if err := s.emailService.SendPasswordResetEmail(
		user.Email,
		resetURL,
	); err != nil {

		// If email delivery fails, revoke the token so that
		// the unused token cannot remain valid.
		_ = s.passwordResetRepository.RevokeActiveByUserID(
			ctx,
			user.ID,
			time.Now(),
		)

		return fmt.Errorf(
			"failed to send password reset email: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Reset Password
// ============================================================

func (s *PasswordResetService) ResetPassword(
	ctx context.Context,
	rawToken string,
	newPassword string,
) error {

	rawToken = strings.TrimSpace(
		rawToken,
	)

	if rawToken == "" {
		return ErrPasswordResetInvalid
	}

	// --------------------------------------------------------
	// A token generated by GenerateSecureToken(32) contains
	// 64 hexadecimal characters.
	//
	// Allow some headroom but reject unreasonable input.
	// --------------------------------------------------------

	if len(rawToken) > 512 {
		return ErrPasswordResetInvalid
	}

	// --------------------------------------------------------
	// Validate password before database work.
	//
	// Your existing password-hashing configuration remains
	// responsible for Argon2id.
	// --------------------------------------------------------

	newPassword = strings.TrimSpace(
		newPassword,
	)

	if newPassword == "" {
		return errors.New(
			"password cannot be empty",
		)
	}

	// --------------------------------------------------------
	// Hash the new password.
	// --------------------------------------------------------

	newPasswordHash, err := s.passwordHasher.Hash(
		newPassword,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to hash new password: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Hash supplied reset token.
	// --------------------------------------------------------

	tokenHash := security.HashToken(
		rawToken,
	)

	// --------------------------------------------------------
	// We need the user ID before the transaction that changes
	// the password.
	//
	// Find the token directly through a small lookup.
	// --------------------------------------------------------

	resetToken, err := s.findResetToken(
		ctx,
		tokenHash,
	)

	if err != nil {
		return ErrPasswordResetInvalid
	}

	// --------------------------------------------------------
	// Transaction performs final token validation again under
	// a row lock, so concurrent reset attempts cannot both
	// succeed.
	// --------------------------------------------------------

	err = s.passwordResetRepository.ResetPassword(
		ctx,
		tokenHash,
		resetToken.UserID,
		newPasswordHash,
		time.Now(),
	)

	if err != nil {

		switch {
		case errors.Is(
			err,
			repository.ErrPasswordResetTokenUsed,
		):
			return ErrPasswordResetInvalid

		case errors.Is(
			err,
			repository.ErrPasswordResetTokenRevoked,
		):
			return ErrPasswordResetInvalid

		case errors.Is(
			err,
			repository.ErrPasswordResetTokenExpired,
		):
			return ErrPasswordResetInvalid

		case errors.Is(
			err,
			gorm.ErrRecordNotFound,
		):
			return ErrPasswordResetInvalid

		default:
			return fmt.Errorf(
				"failed to reset password: %w",
				err,
			)
		}
	}

	return nil
}

func (s *PasswordResetService) findResetToken(
	ctx context.Context,
	tokenHash string,
) (*models.PasswordResetToken, error) {

	token, err := s.passwordResetRepository.FindByTokenHash(
		ctx,
		tokenHash,
	)

	if err != nil {
		return nil, err
	}

	now := time.Now()

	if token.UsedAt != nil {
		return nil, ErrPasswordResetInvalid
	}

	if token.RevokedAt != nil {
		return nil, ErrPasswordResetInvalid
	}

	if !now.Before(token.ExpiresAt) {
		return nil, ErrPasswordResetInvalid
	}

	return token, nil
}
