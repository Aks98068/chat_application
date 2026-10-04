package handlres

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Aks98068/forensics/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authservice *service.AuthService
}

func NewAuthHandler(authservice *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authservice: authservice,
	}
}

type ForgotPasswordRequest struct {
	Identifier string `json:"identifier"`
}
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}
type RegisterRequest struct {
	Email string `json:"email" binding:"required,email"`

	Username string `json:"username" binding:"required,min=3,max=50"`

	Password string `json:"password" binding:"required,min=8,max=128"`

	FirstName string `json:"firstName" binding:"required,min=1,max=100"`

	LastName string `json:"lastName" binding:"required,min=1,max=100"`
}
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required,min=3,max=255"`
	Password   string `json:"password" binding:"required,min=8,max=128"`
}
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {

	// 1. Parse and validate JSON request

	var request RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid registration data",
		})
		return
	}

	// 2. Basic whitespace validation

	if strings.TrimSpace(request.Email) == "" ||
		strings.TrimSpace(request.Username) == "" ||
		strings.TrimSpace(request.FirstName) == "" ||
		strings.TrimSpace(request.LastName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "registration fields cannot be empty",
		})
		return
	}

	input := service.RegisterInput{
		Email:     request.Email,
		Username:  request.Username,
		Password:  request.Password,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	}

	// Call service

	_, err := h.authservice.Register(
		c.Request.Context(),
		input,
	)

	if err != nil {

		// Duplicate username

		if errors.Is(err, service.ErrUsernameExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "username already exists",
			})
			return
		}

		// duplicate email

		if errors.Is(err, service.ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "email already exists",
			})
			return
		}

		if strings.Contains(err.Error(), "required") ||
			strings.Contains(err.Error(), "must be") ||
			strings.Contains(err.Error(), "can contain") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid registration data",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "registration failed",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registration successful",
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var request VerifyEmailRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "verification token is required",
			},
		)
		return
	}

	err := h.authservice.VerifyEmail(
		c.Request.Context(),
		request.Token,
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidVerification) ||
			errors.Is(err, service.ErrVerificationUsed) {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "invalid or expired verification token",
				},
			)
			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "email verification failed",
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "email verified successfully",
		},
	)
}

func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	var request ResendVerificationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "valid email is required",
			},
		)
		return
	}

	err := h.authservice.ResendVerificationEmail(
		c.Request.Context(),
		request.Email,
	)

	if err != nil {
		if errors.Is(err, service.ErrResendTooSoon) {
			// Do not expose exact account/token state.
			c.JSON(
				http.StatusTooManyRequests,
				gin.H{
					"error": "please wait before requesting another verification email",
				},
			)
			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "unable to process verification email request",
			},
		)
		return
	}

	// Generic response intentionally.
	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "if the account requires verification, a verification email has been sent",
		},
	)
}

//
//
//

func (h *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid login data",
		})
		return
	}

	request.Identifier = strings.TrimSpace(
		request.Identifier,
	)

	if request.Identifier == "" ||
		request.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "identifier and password are required",
		})
		return
	}

	input := service.LoginInput{
		Identifier: request.Identifier,
		Password:   request.Password,
	}

	result, err := h.authservice.Login(
		c.Request.Context(),
		input,
		c.Request.UserAgent(),
		c.ClientIP(),
	)

	if err != nil {

		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid username/email or password",
			})
			return
		}

		if errors.Is(err, service.ErrAccountSuspended) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "account is suspended",
			})
			return
		}

		if errors.Is(err, service.ErrAccountDisabled) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "account is disabled",
			})
			return
		}

		if errors.Is(err, service.ErrEmailNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "email address is not verified",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "login failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",

		"accessToken": result.Tokens.AccessToken,

		"refreshToken": result.Tokens.RefreshToken,

		"accessTokenExpiresIn":  result.Tokens.AccessTokenExpiresIn,
		"refreshTokenExpiresIn": result.Tokens.RefreshTokenExpiresIn,

		"user": gin.H{
			"id":       result.User.ID,
			"username": result.User.Username,
			"email":    result.User.Email,
			"role":     result.User.Role,
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {

	var request RefreshTokenRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "refresh token is required",
		})
		return
	}

	request.RefreshToken = strings.TrimSpace(request.RefreshToken)

	if request.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "refresh token is required",
		})
		return
	}

	tokens, err := h.authservice.RefreshTokens(
		c.Request.Context(),
		request.RefreshToken,
		c.Request.UserAgent(),
		c.ClientIP(),
	)

	if err != nil {

		if errors.Is(err, service.ErrInvalidRefreshToken) ||
			errors.Is(err, service.ErrRefreshTokenExpired) ||
			errors.Is(err, service.ErrRefreshTokenRevoked) {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid refresh token",
			})
			return
		}

		if errors.Is(err, service.ErrRefreshTokenReuse) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "refresh token reuse detected",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to refresh authentication",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "token refreshed successfully",

		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,

		"accessTokenExpiresIn":  tokens.AccessTokenExpiresIn,
		"refreshTokenExpiresIn": tokens.RefreshTokenExpiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {

	var request LogoutRequest

	// --------------------------------------------------------
	// Parse request body.
	// --------------------------------------------------------

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request body",
			},
		)

		return
	}

	// --------------------------------------------------------
	// Normalize token.
	// --------------------------------------------------------

	request.RefreshToken = strings.TrimSpace(
		request.RefreshToken,
	)

	// --------------------------------------------------------
	// Logout is idempotent.
	//
	// We still reject a completely missing request field at
	// the HTTP boundary because the client has malformed input.
	// --------------------------------------------------------

	if request.RefreshToken == "" {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "refreshToken is required",
			},
		)

		return
	}

	// --------------------------------------------------------
	// Revoke refresh token.
	// --------------------------------------------------------

	if err := h.authservice.Logout(
		c.Request.Context(),
		request.RefreshToken,
	); err != nil {

		// ----------------------------------------------------
		// Do not expose database or internal errors.
		// ----------------------------------------------------

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "failed to logout",
			},
		)

		return
	}

	// --------------------------------------------------------
	// Successful logout.
	// --------------------------------------------------------

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "logged out successfully",
		},
	)
}

func (h *AuthHandler) ForgotPassword(
	c *gin.Context,
) {

	var request ForgotPasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request body",
			},
		)

		return
	}

	request.Identifier = strings.TrimSpace(
		request.Identifier,
	)

	if request.Identifier == "" {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "identifier is required",
			},
		)

		return
	}

	err := h.authservice.RequestPasswordReset(
		c.Request.Context(),
		request.Identifier,
	)

	if err != nil {

		// --------------------------------------------
		// Do not reveal whether the account exists.
		// --------------------------------------------

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "unable to process password reset request",
			},
		)

		return
	}

	// ------------------------------------------------
	// IMPORTANT:
	//
	// This response is the same whether the account
	// exists or does not exist.
	// ------------------------------------------------

	c.JSON(
		http.StatusAccepted,
		gin.H{
			"message": "if an account matches the provided identifier, a password reset email has been sent",
		},
	)
}

func (h *AuthHandler) ResetPassword(
	c *gin.Context,
) {

	var request ResetPasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request body",
			},
		)

		return
	}

	request.Token = strings.TrimSpace(
		request.Token,
	)

	if request.Token == "" {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "reset token is required",
			},
		)

		return
	}

	if request.NewPassword == "" {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "new password is required",
			},
		)

		return
	}

	err := h.authservice.ResetPassword(
		c.Request.Context(),
		request.Token,
		request.NewPassword,
	)

	if err != nil {

		if errors.Is(
			err,
			service.ErrPasswordResetInvalid,
		) {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "invalid or expired password reset token",
				},
			)

			return
		}

		if strings.Contains(
			err.Error(),
			"password",
		) {

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "failed to reset password",
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "password reset successfully",
		},
	)
}
