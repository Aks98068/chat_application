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

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(
	db *gorm.DB,
) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

// ============================================================
// Rotation Result
// ============================================================

type RefreshTokenRotationResult struct {
	OldToken      *models.RefreshToken
	NewToken      *models.RefreshToken
	ReuseDetected bool
	TokenExpired  bool
	TokenRevoked  bool
}

// ============================================================
// Create
// ============================================================

func (r *RefreshTokenRepository) Create(
	ctx context.Context,
	token *models.RefreshToken,
) error {

	return r.db.
		WithContext(ctx).
		Create(token).
		Error
}

// ============================================================
// Save
// ============================================================

func (r *RefreshTokenRepository) Save(
	ctx context.Context,
	token *models.RefreshToken,
) error {

	return r.db.
		WithContext(ctx).
		Save(token).
		Error
}

// ============================================================
// Find By Token Hash
// ============================================================

func (r *RefreshTokenRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*models.RefreshToken, error) {

	var token models.RefreshToken

	err := r.db.
		WithContext(ctx).
		Preload("User").
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

// ============================================================
// Revoke
// ============================================================

func (r *RefreshTokenRepository) Revoke(
	ctx context.Context,
	tokenID uuid.UUID,
	revokedAt time.Time,
) error {

	result := r.db.
		WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where(
			"id = ? AND revoked_at IS NULL",
			tokenID,
		).
		Update(
			"revoked_at",
			revokedAt,
		)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// ============================================================
// Revoke By Token Hash
// ============================================================
//
// Used by logout.
//
// This performs the lookup and revocation in one database
// operation rather than:
//
// 1. SELECT token
// 2. UPDATE token
//
// That avoids an unnecessary race window.
//
// ============================================================

func (r *RefreshTokenRepository) RevokeByTokenHash(
	ctx context.Context,
	tokenHash string,
	revokedAt time.Time,
) error {

	if tokenHash == "" {
		return errors.New("token hash cannot be empty")
	}

	result := r.db.
		WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where(
			"token_hash = ? AND revoked_at IS NULL",
			tokenHash,
		).
		Update(
			"revoked_at",
			revokedAt,
		)

	if result.Error != nil {
		return result.Error
	}

	// We deliberately do NOT return gorm.ErrRecordNotFound
	// when no row was changed.
	//
	// Logout is intentionally idempotent.
	//
	// Therefore:
	//
	// invalid token
	// expired token
	// already revoked token
	//
	// all result in the same logout behavior.
	//
	// This also avoids exposing whether a refresh token exists.

	return nil
}

// ============================================================
// Set Replaced By
// ============================================================

func (r *RefreshTokenRepository) SetReplacedBy(
	ctx context.Context,
	tokenID uuid.UUID,
	replacedByTokenID uuid.UUID,
) error {

	result := r.db.
		WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ?", tokenID).
		Update(
			"replaced_by_token_id",
			replacedByTokenID,
		)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// ============================================================
// Revoke Token Family
// ============================================================

func (r *RefreshTokenRepository) RevokeTokenFamily(
	ctx context.Context,
	tokenFamily uuid.UUID,
	revokedAt time.Time,
) error {

	return r.db.
		WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where(
			"token_family = ? AND revoked_at IS NULL",
			tokenFamily,
		).
		Update(
			"revoked_at",
			revokedAt,
		).
		Error
}

// ============================================================
// Rotate Refresh Token
// ============================================================

func (r *RefreshTokenRepository) Rotate(
	ctx context.Context,
	tokenHash string,
	newToken *models.RefreshToken,
	now time.Time,
) (*RefreshTokenRotationResult, error) {

	var result RefreshTokenRotationResult

	err := r.db.
		WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {

			var oldToken models.RefreshToken

			err := tx.
				Clauses(
					clause.Locking{
						Strength: "UPDATE",
					},
				).
				Preload("User").
				Where(
					"token_hash = ?",
					tokenHash,
				).
				First(&oldToken).
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

			result.OldToken = &oldToken

			// ------------------------------------------------
			// Check expiration.
			// ------------------------------------------------

			if now.After(oldToken.ExpiresAt) {
				result.TokenExpired = true
				return nil
			}

			// ------------------------------------------------
			// Check revocation.
			// ------------------------------------------------

			if oldToken.RevokedAt != nil {

				// --------------------------------------------
				// Reuse detection.
				// --------------------------------------------

				if oldToken.ReplacedByTokenID != nil {

					result.ReuseDetected = true

					if err := tx.
						Model(&models.RefreshToken{}).
						Where(
							"token_family = ? AND revoked_at IS NULL",
							oldToken.TokenFamily,
						).
						Update(
							"revoked_at",
							now,
						).
						Error; err != nil {
						return err
					}

					return nil
				}

				result.TokenRevoked = true
				return nil
			}

			// ------------------------------------------------
			// Preserve token family.
			// ------------------------------------------------

			newToken.ID = uuid.New()
			newToken.UserID = oldToken.UserID
			newToken.TokenFamily = oldToken.TokenFamily

			// ------------------------------------------------
			// Create new refresh token.
			// ------------------------------------------------

			if err := tx.
				Create(newToken).
				Error; err != nil {
				return err
			}

			// ------------------------------------------------
			// Revoke old refresh token.
			// ------------------------------------------------

			updateResult := tx.
				Model(&models.RefreshToken{}).
				Where(
					"id = ? AND revoked_at IS NULL",
					oldToken.ID,
				).
				Update(
					"revoked_at",
					now,
				)

			if updateResult.Error != nil {
				return updateResult.Error
			}

			if updateResult.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}

			// ------------------------------------------------
			// Link old token to replacement token.
			// ------------------------------------------------

			updateResult = tx.
				Model(&models.RefreshToken{}).
				Where(
					"id = ?",
					oldToken.ID,
				).
				Update(
					"replaced_by_token_id",
					newToken.ID,
				)

			if updateResult.Error != nil {
				return updateResult.Error
			}

			if updateResult.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}

			result.NewToken = newToken

			return nil
		})

	if err != nil {
		return nil, err
	}

	return &result, nil
}
