package service

import (
	"context"
	"errors"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/Aks98068/forensics/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(
	userRepository *repository.UserRepository,
) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

type AuthenticatedUser struct {
	ID            uuid.UUID         `json:"id"`
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	FirstName     string            `json:"firstName"`
	LastName      string            `json:"lastName"`
	Role          models.UserRole   `json:"role"`
	Status        models.UserStatus `json:"status"`
	EmailVerified bool              `json:"emailVerified"`
}

func (s *UserService) GetAuthenticatedUser(
	ctx context.Context,
	userID uuid.UUID,
) (*AuthenticatedUser, error) {

	if userID == uuid.Nil {
		return nil, errors.New("invalid authenticated user ID")
	}

	user, err := s.userRepository.FindByID(
		ctx,
		userID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, err
	}

	return &AuthenticatedUser{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
	}, nil
}
