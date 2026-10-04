package repository

import (
	"context"
	"errors"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// ============================================================
// Find user by username
// ============================================================

func (r *UserRepository) FindByUsername(
	ctx context.Context,
	username string,
) (*models.User, error) {

	var user models.User

	err := r.db.
		WithContext(ctx).
		Where("username = ?", username).
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

// ============================================================
// Find user by email
// ============================================================

func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {

	var user models.User

	err := r.db.
		WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

// ============================================================
// Find user by ID
// ============================================================
//
// Used by authenticated endpoints such as:
//
// GET /api/v1/protected/user/me
//
// The ID comes from the verified JWT and is then used to
// retrieve the current user from the database.
//
// ============================================================

func (r *UserRepository) FindByID(
	ctx context.Context,
	userID uuid.UUID,
) (*models.User, error) {

	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be nil")
	}

	var user models.User

	err := r.db.
		WithContext(ctx).
		Where("id = ?", userID).
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

// ============================================================
// Create user
// ============================================================

func (r *UserRepository) Create(
	ctx context.Context,
	user *models.User,
) error {

	return r.db.
		WithContext(ctx).
		Create(user).
		Error
}

// ============================================================
// Mark email as verified
// ============================================================

func (r *UserRepository) MarkEmailAsVerified(
	ctx context.Context,
	userID uuid.UUID,
) error {

	if userID == uuid.Nil {
		return errors.New("user ID cannot be nil")
	}

	result := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update(
			"email_verified",
			true,
		)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *UserRepository) FindByIdentifier(
	ctx context.Context,
	identifier string,
) (*models.User, error) {

	var user models.User

	err := r.db.
		WithContext(ctx).
		Where(
			"username = ? OR email = ?",
			identifier,
			identifier,
		).
		First(&user).
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

	return &user, nil
}
