package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/Aks98068/forensics/internal/repository"
	"github.com/Aks98068/forensics/internal/security"

	"gorm.io/gorm"
)

type RegisterInput struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
}

type LoginInput struct {
	Identifier string
	Password   string
}

type AuthService struct {
	userRepository           *repository.UserRepository
	passwordhasher           *security.PasswordHash
	emailverificationservice *EmailVerificationService
	emailservice             *EmailService
	authtokenservice         *AuthTokenService
	passwordResetService     *PasswordResetService
	appURL                   string
}
type LoginResult struct {
	User   *models.User
	Tokens *AuthenticationTokens
}

func (s *AuthService) ResendVerificationEmail(
	ctx context.Context,
	email string,
) error {
	return s.emailverificationservice.ResendVerificationEmail(
		ctx,
		email,
	)
}
func NewAuthService(
	userRepository *repository.UserRepository,
	passwordhasher *security.PasswordHash,
	emailverificationservice *EmailVerificationService,
	emailservice *EmailService,
	authtokenservice *AuthTokenService,
	appURL string,
	passwordResetService *PasswordResetService,
) *AuthService {
	return &AuthService{
		userRepository:           userRepository,
		passwordhasher:           passwordhasher,
		emailverificationservice: emailverificationservice,
		emailservice:             emailservice,
		authtokenservice:         authtokenservice,
		passwordResetService:     passwordResetService,
		appURL:                   strings.TrimRight(appURL, "/"),
	}
}

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

func (s *AuthService) Register(
	ctx context.Context,
	input RegisterInput,
) (*models.User, error) {

	// ------------------------------------------------------------
	// 1. Normalize input
	// ------------------------------------------------------------

	email := strings.ToLower(strings.TrimSpace(input.Email))
	username := strings.ToLower(strings.TrimSpace(input.Username))

	firstName := strings.TrimSpace(input.FirstName)
	lastName := strings.TrimSpace(input.LastName)

	// ------------------------------------------------------------
	// 2. Validate normalized input
	// ------------------------------------------------------------

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if !usernamePattern.MatchString(username) {
		return nil, fmt.Errorf(
			"username can contain only lowercase letters, numbers, and underscores",
		)
	}

	if utf8.RuneCountInString(username) < 3 ||
		utf8.RuneCountInString(username) > 50 {
		return nil, fmt.Errorf(
			"username must be between 3 and 50 characters",
		)
	}

	if utf8.RuneCountInString(email) > 255 {
		return nil, fmt.Errorf(
			"email must not exceed 255 characters",
		)
	}

	if utf8.RuneCountInString(input.Password) < 8 ||
		utf8.RuneCountInString(input.Password) > 128 {
		return nil, fmt.Errorf(
			"password must be between 8 and 128 characters",
		)
	}

	if firstName == "" {
		return nil, fmt.Errorf("first name is required")
	}

	if lastName == "" {
		return nil, fmt.Errorf("last name is required")
	}

	if utf8.RuneCountInString(firstName) > 100 {
		return nil, fmt.Errorf(
			"first name must not exceed 100 characters",
		)
	}

	if utf8.RuneCountInString(lastName) > 100 {
		return nil, fmt.Errorf(
			"last name must not exceed 100 characters",
		)
	}

	// ------------------------------------------------------------
	// 3. Check username uniqueness
	// ------------------------------------------------------------

	existingUser, err := s.userRepository.FindByUsername(
		ctx,
		username,
	)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	// ------------------------------------------------------------
	// 4. Check email uniqueness
	// ------------------------------------------------------------

	existingUser, err = s.userRepository.FindByEmail(
		ctx,
		email,
	)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// ------------------------------------------------------------
	// 5. Hash password
	// ------------------------------------------------------------

	passwordHash, err := s.passwordhasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 6. Create user
	// ------------------------------------------------------------

	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,

		// Never accept role from public registration.
		Role: models.RoleUser,

		// Never accept status from public registration.
		Status: models.StatusActive,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 7. Create email verification token
	// ------------------------------------------------------------

	verificationToken, err := s.emailverificationservice.CreateToken(
		ctx,
		user.ID,
	)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 8. Build verification URL
	// ------------------------------------------------------------

	verificationURL := s.appURL +
		"/verify-email?token=" +
		verificationToken

	// ------------------------------------------------------------
	// 9. Send verification email
	// ------------------------------------------------------------

	if err := s.emailservice.SendVerificationEmail(
		user.Email,
		verificationURL,
	); err != nil {
		return nil, err
	}

	return user, nil
}

// ------------------------------------------------------------
// Verify Email
// ------------------------------------------------------------

func (s *AuthService) VerifyEmail(
	ctx context.Context,
	rawToken string,
) error {
	return s.emailverificationservice.VerifyEmail(
		ctx,
		rawToken,
	)
}

func (s *AuthService) Login(
	ctx context.Context,
	input LoginInput,
	userAgent string,
	ipAddress string,
) (*LoginResult, error) {

	identifier := strings.ToLower(
		strings.TrimSpace(input.Identifier),
	)

	if identifier == "" {
		return nil, ErrInvalidCredentials
	}

	if input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	var user *models.User
	var err error

	if strings.Contains(identifier, "@") {

		user, err = s.userRepository.FindByEmail(
			ctx,
			identifier,
		)

	} else {

		user, err = s.userRepository.FindByUsername(
			ctx,
			identifier,
		)
	}

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	passwordCorrect, err := s.passwordhasher.Verify(
		input.Password,
		user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	if !passwordCorrect {
		return nil, ErrInvalidCredentials
	}

	if user.Status == models.StatusSuspend {
		return nil, ErrAccountSuspended
	}

	if user.Status == models.StatusDisable {
		return nil, ErrAccountDisabled
	}

	if !user.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	tokens, err := s.authtokenservice.CreateLoginTokens(
		ctx,
		user,
		userAgent,
		ipAddress,
	)

	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *AuthService) RefreshTokens(
	ctx context.Context,
	rawRefreshToken string,
	userAgent string,
	ipAddress string,
) (*AuthenticationTokens, error) {

	return s.authtokenservice.Refresh(
		ctx,
		rawRefreshToken,
		userAgent,
		ipAddress,
	)
}

// ============================================================
// Logout
// ============================================================

func (s *AuthService) Logout(
	ctx context.Context,
	rawRefreshToken string,
) error {

	return s.authtokenservice.Logout(
		ctx,
		rawRefreshToken,
	)
}

func (s *AuthService) RequestPasswordReset(
	ctx context.Context,
	identifier string,
) error {

	return s.passwordResetService.RequestPasswordReset(
		ctx,
		identifier,
	)
}
func (s *AuthService) ResetPassword(
	ctx context.Context,
	token string,
	newPassword string,
) error {

	return s.passwordResetService.ResetPassword(
		ctx,
		token,
		newPassword,
	)
}
