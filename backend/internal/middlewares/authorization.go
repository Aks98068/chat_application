package middlewares

import (
	"errors"
	"net/http"

	"github.com/Aks98068/forensics/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================================
// Errors
// ============================================================

var (
	ErrUserIDNotFound = errors.New(
		"authenticated user ID not found in context",
	)

	ErrRoleNotFound = errors.New(
		"authenticated user role not found in context",
	)
)

// ============================================================
// Require Roles
// ============================================================
//
// RequireRoles checks whether the authenticated user's role
// is allowed to access the route.
//
// Example:
//
// RequireRoles(
//     models.RoleAdmin,
// )
//
// or:
//
// RequireRoles(
//     models.RoleAnalyst,
//     models.RoleAdmin,
// )
//
// IMPORTANT:
// JWTAuthMiddleware must run before RequireRoles.
//
// ============================================================

func RequireRoles(
	allowedRoles ...models.UserRole,
) gin.HandlerFunc {

	// --------------------------------------------------------
	// Build a set of allowed roles.
	//
	// A map gives us constant-time role lookup.
	// --------------------------------------------------------

	allowedRoleSet := make(
		map[models.UserRole]struct{},
		len(allowedRoles),
	)

	for _, role := range allowedRoles {

		// ----------------------------------------------------
		// Make sure the application developer did not
		// accidentally configure an invalid role.
		// ----------------------------------------------------

		if !isValidRole(role) {
			panic("invalid role supplied to RequireRoles")
		}

		allowedRoleSet[role] = struct{}{}
	}

	// --------------------------------------------------------
	// Require at least one role.
	// --------------------------------------------------------

	if len(allowedRoleSet) == 0 {
		panic("RequireRoles requires at least one role")
	}

	// --------------------------------------------------------
	// Return the actual Gin middleware.
	// --------------------------------------------------------

	return func(c *gin.Context) {

		// ----------------------------------------------------
		// Get the role placed into the Gin context by
		// JWTAuthMiddleware.
		// ----------------------------------------------------

		roleValue, exists := c.Get(ContextRole)

		if !exists {
			forbidden(c)
			return
		}

		// ----------------------------------------------------
		// Make sure the context value is actually a
		// models.UserRole.
		// ----------------------------------------------------

		role, ok := roleValue.(models.UserRole)

		if !ok {
			forbidden(c)
			return
		}

		// ----------------------------------------------------
		// Make sure the role itself is valid.
		// ----------------------------------------------------

		if !isValidRole(role) {
			forbidden(c)
			return
		}

		// ----------------------------------------------------
		// Check whether the user's role is allowed.
		// ----------------------------------------------------

		if _, allowed := allowedRoleSet[role]; !allowed {
			forbidden(c)
			return
		}

		// ----------------------------------------------------
		// Authorization successful.
		// Continue to the next handler/middleware.
		// ----------------------------------------------------

		c.Next()
	}
}

// ============================================================
// Get Authenticated User ID
// ============================================================

func GetAuthenticatedUserID(
	c *gin.Context,
) (uuid.UUID, error) {

	value, exists := c.Get(ContextUserID)

	if !exists {
		return uuid.Nil, ErrUserIDNotFound
	}

	userID, ok := value.(uuid.UUID)

	if !ok || userID == uuid.Nil {
		return uuid.Nil, ErrUserIDNotFound
	}

	return userID, nil
}

// ============================================================
// Get Authenticated Username
// ============================================================

func GetAuthenticatedUsername(
	c *gin.Context,
) (string, error) {

	value, exists := c.Get(ContextUsername)

	if !exists {
		return "", errors.New(
			"authenticated username not found in context",
		)
	}

	username, ok := value.(string)

	if !ok || username == "" {
		return "", errors.New(
			"authenticated username not found in context",
		)
	}

	return username, nil
}

// ============================================================
// Get Authenticated Role
// ============================================================

func GetAuthenticatedRole(
	c *gin.Context,
) (models.UserRole, error) {

	value, exists := c.Get(ContextRole)

	if !exists {
		return "", ErrRoleNotFound
	}

	role, ok := value.(models.UserRole)

	if !ok || !isValidRole(role) {
		return "", ErrRoleNotFound
	}

	return role, nil
}

// ============================================================
// Forbidden Response
// ============================================================

func forbidden(c *gin.Context) {

	c.JSON(
		http.StatusForbidden,
		gin.H{
			"error": "forbidden",
		},
	)

	c.Abort()
}
