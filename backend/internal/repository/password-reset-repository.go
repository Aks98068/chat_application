package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrPasswordResetTokenUsed = errors.New(
		"password reset token already used",
	)

	ErrPasswordResetTokenRevoked = errors.New(
		"password reset token revoked",
	)

	ErrPasswordResetTokenExpired = errors.New(
		"password reset token expired",
	)
)

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(
	db *gorm.DB,
) *PasswordResetRepository {
	return &PasswordResetRepository{
		db: db,
	}
}

// ============================================================
// Create
// ============================================================

func (r *PasswordResetRepository) Create(
	ctx context.Context,
	token *models.PasswordResetToken,
) error {

	return r.db.
		WithContext(ctx).
		Create(token).
		Error
}

// ============================================================
// Revoke active reset tokens for a user
// ============================================================
//
// This guarantees that requesting a new password reset
// invalidates previous reset links.
//
// ============================================================

func (r *PasswordResetRepository) RevokeActiveByUserID(
	ctx context.Context,
	userID uuid.UUID,
	revokedAt time.Time,
) error {

	if userID == uuid.Nil {
		return errors.New("user ID cannot be nil")
	}

	return r.db.
		WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where(
			"user_id = ? AND used_at IS NULL AND revoked_at IS NULL",
			userID,
		).
		Update(
			"revoked_at",
			revokedAt,
		).
		Error
}

// ============================================================
// Reset password transaction
// ============================================================
//
// This performs:
//
// 1. Lock reset token.
// 2. Validate token state.
// 3. Update password.
// 4. Mark token as used.
// 5. Revoke all other reset tokens.
// 6. Revoke all refresh sessions.
//
// Everything succeeds or everything rolls back.
//
// ============================================================

func (r *PasswordResetRepository) ResetPassword(
	ctx context.Context,
	tokenHash string,
	userID uuid.UUID,
	passwordHash string,
	now time.Time,
) error {

	if tokenHash == "" {
		return errors.New("token hash cannot be empty")
	}

	if userID == uuid.Nil {
		return errors.New("user ID cannot be nil")
	}

	if passwordHash == "" {
		return errors.New("password hash cannot be empty")
	}

	return r.db.
		WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {

			var resetToken models.PasswordResetToken

			// ------------------------------------------------
			// Lock the reset-token row.
			//
			// This prevents two simultaneous reset requests
			// from consuming the same token.
			// ------------------------------------------------

			err := tx.
				Clauses(
					clause.Locking{
						Strength: "UPDATE",
					},
				).
				Where(
					"token_hash = ?",
					tokenHash,
				).
				First(&resetToken).
				Error

			if err != nil {

				if errors.Is(
					err,
					gorm.ErrRecordNotFound,
				) {
					return gorm.ErrRecordNotFound
				}

				return err
			}

			// ------------------------------------------------
			// Make sure the token belongs to the expected user.
			// ------------------------------------------------

			if resetToken.UserID != userID {
				return gorm.ErrRecordNotFound
			}

			// ------------------------------------------------
			// Check whether the token has already been used.
			// ------------------------------------------------

			if resetToken.UsedAt != nil {
				return ErrPasswordResetTokenUsed
			}

			// ------------------------------------------------
			// Check whether the token was revoked.
			// ------------------------------------------------

			if resetToken.RevokedAt != nil {
				return ErrPasswordResetTokenRevoked
			}

			// ------------------------------------------------
			// Check expiration.
			// ------------------------------------------------

			if !now.Before(resetToken.ExpiresAt) {
				return ErrPasswordResetTokenExpired
			}

			// ------------------------------------------------
			// Verify the user still exists.
			// ------------------------------------------------

			var user models.User

			err = tx.
				Where(
					"id = ?",
					userID,
				).
				First(&user).
				Error

			if err != nil {

				if errors.Is(
					err,
					gorm.ErrRecordNotFound,
				) {
					return gorm.ErrRecordNotFound
				}

				return err
			}

			// ------------------------------------------------
			// Update password.
			// ------------------------------------------------

			updateResult := tx.
				Model(&models.User{}).
				Where(
					"id = ?",
					userID,
				).
				Update(
					"password_hash",
					passwordHash,
				)

			if updateResult.Error != nil {
				return updateResult.Error
			}

			if updateResult.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}

			// ------------------------------------------------
			// Mark current reset token as used.
			// ------------------------------------------------

			updateResult = tx.
				Model(&models.PasswordResetToken{}).
				Where(
					"id = ? AND used_at IS NULL AND revoked_at IS NULL",
					resetToken.ID,
				).
				Update(
					"used_at",
					now,
				)

			if updateResult.Error != nil {
				return updateResult.Error
			}

			if updateResult.RowsAffected != 1 {
				return errors.New(
					"password reset token could not be consumed",
				)
			}

			// ------------------------------------------------
			// Revoke every other active reset token.
			// ------------------------------------------------

			updateResult = tx.
				Model(&models.PasswordResetToken{}).
				Where(
					"user_id = ? AND id <> ? AND used_at IS NULL AND revoked_at IS NULL",
					userID,
					resetToken.ID,
				).
				Update(
					"revoked_at",
					now,
				)

			if updateResult.Error != nil {
				return updateResult.Error
			}

			// ------------------------------------------------
			// Revoke every active refresh-token session.
			//
			// This is extremely important.
			//
			// If an attacker somehow has an existing refresh
			// token, changing the password should terminate
			// that session.
			// ------------------------------------------------

			updateResult = tx.
				Model(&models.RefreshToken{}).
				Where(
					"user_id = ? AND revoked_at IS NULL",
					userID,
				).
				Update(
					"revoked_at",
					now,
				)

			if updateResult.Error != nil {
				return updateResult.Error
			}

			return nil
		})
}

func (r *PasswordResetRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*models.PasswordResetToken, error) {

	if tokenHash == "" {
		return nil, errors.New(
			"token hash cannot be empty",
		)
	}

	var token models.PasswordResetToken

	err := r.db.
		WithContext(ctx).
		Where(
			"token_hash = ?",
			tokenHash,
		).
		First(&token).
		Error

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &token, nil
}
