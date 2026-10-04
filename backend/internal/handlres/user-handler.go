package handlres

import (
	"errors"
	"net/http"

	middleware "github.com/Aks98068/forensics/internal/middlewares"
	"github.com/Aks98068/forensics/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================================
// User Handler
// ============================================================

type UserHandler struct {
	userService *service.UserService
}

// ============================================================
// Constructor
// ============================================================

func NewUserHandler(
	userService *service.UserService,
) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ============================================================
// GET CURRENT AUTHENTICATED USER
// ============================================================
//
// GET /api/v1/protected/user/me
//
// Middleware:
//
// 1. JWTAuthMiddleware
// 2. RequireRoles(USER, ANALYST, ADMIN)
//
// Flow:
//
// Request
//   ↓
// JWT authentication
//   ↓
// User ID extracted from JWT
//   ↓
// UserService
//   ↓
// UserRepository
//   ↓
// MySQL
//   ↓
// Safe user response
//
// ============================================================

func (h *UserHandler) Me(c *gin.Context) {

	// --------------------------------------------------------
	// Get authenticated user ID from JWT middleware.
	// --------------------------------------------------------

	userID, err := middleware.GetAuthenticatedUserID(c)

	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthorized",
			},
		)

		c.Abort()
		return
	}

	// --------------------------------------------------------
	// Retrieve the current user from the database.
	// --------------------------------------------------------

	user, err := h.userService.GetAuthenticatedUser(
		c.Request.Context(),
		userID,
	)

	if err != nil {

		// ----------------------------------------------------
		// The authenticated user no longer exists.
		// ----------------------------------------------------

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "unauthorized",
				},
			)

			c.Abort()
			return
		}

		// ----------------------------------------------------
		// Unexpected service/database error.
		// ----------------------------------------------------

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "failed to retrieve authenticated user",
			},
		)

		c.Abort()
		return
	}

	// --------------------------------------------------------
	// Return the safe user representation.
	//
	// PasswordHash is NOT included because the service
	// returns AuthenticatedUser rather than models.User.
	// --------------------------------------------------------

	c.JSON(
		http.StatusOK,
		gin.H{
			"user": user,
		},
	)
}
