package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(
	db *gorm.DB,
) *EmailVerificationRepository {
	return &EmailVerificationRepository{
		db: db,
	}
}

// ------------------------------------------------------------
// Create token
// ------------------------------------------------------------

func (r *EmailVerificationRepository) Create(
	ctx context.Context,
	token *models.EmailVerificationToken,
) error {
	return r.db.WithContext(ctx).
		Create(token).
		Error
}

// ------------------------------------------------------------
// Find token by hash
// ------------------------------------------------------------

func (r *EmailVerificationRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*models.EmailVerificationToken, error) {

	var token models.EmailVerificationToken

	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&token).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &token, nil
}

// ------------------------------------------------------------
// Mark token as used
// ------------------------------------------------------------

func (r *EmailVerificationRepository) MarkAsUsed(
	ctx context.Context,
	tokenID uuid.UUID,
	usedAt time.Time,
) error {

	result := r.db.WithContext(ctx).
		Model(&models.EmailVerificationToken{}).
		Where(
			"id = ? AND used_at IS NULL AND revoked_at IS NULL",
			tokenID,
		).
		Update("used_at", usedAt)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// ------------------------------------------------------------
// Revoke active verification tokens
// ------------------------------------------------------------

func (r *EmailVerificationRepository) RevokeActiveTokens(
	ctx context.Context,
	userID uuid.UUID,
	revokedAt time.Time,
) error {

	return r.db.WithContext(ctx).
		Model(&models.EmailVerificationToken{}).
		Where(
			"user_id = ? AND used_at IS NULL AND revoked_at IS NULL",
			userID,
		).
		Update("revoked_at", revokedAt).
		Error
}

// ------------------------------------------------------------
// Find most recent token
// ------------------------------------------------------------

func (r *EmailVerificationRepository) FindLatestToken(
	ctx context.Context,
	userID uuid.UUID,
) (*models.EmailVerificationToken, error) {

	var token models.EmailVerificationToken

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&token).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &token, nil
}
