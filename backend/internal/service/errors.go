package service

import "errors"

var (
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")

	ErrInvalidVerification = errors.New("invalid or expired verification token")
	ErrVerificationUsed    = errors.New("verification token already used")
	ErrVerificationRevoked = errors.New("verification token revoked")
	ErrResendTooSoon       = errors.New("verification email resend requested too soon")

	ErrInvalidCredentials = errors.New("invalid username/email or password")
	ErrAccountSuspended   = errors.New("account is suspended")
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrEmailNotVerified   = errors.New("email address is not verified")

	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")
	ErrRefreshTokenReuse   = errors.New("refresh token reuse detected")
)
