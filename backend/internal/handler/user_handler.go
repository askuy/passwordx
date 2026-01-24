package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/askuy/passwordx/backend/internal/middleware"
	"github.com/askuy/passwordx/backend/internal/model"
	"github.com/askuy/passwordx/backend/internal/repository"
	"github.com/askuy/passwordx/backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	userRepo    *repository.UserRepository
	tenantRepo  *repository.TenantRepository
}

func NewUserHandler(userService *service.UserService, userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository) *UserHandler {
	return &UserHandler{
		userService: userService,
		userRepo:    userRepo,
		tenantRepo:  tenantRepo,
	}
}

// getCurrentUser fetches the current user from the database using the JWT claims
func (h *UserHandler) getCurrentUser(c *gin.Context) (*model.User, error) {
	userID := middleware.GetUserID(c)
	return h.userRepo.GetByID(c.Request.Context(), userID)
}

// Create creates a new user (admin only)
func (h *UserHandler) Create(c *gin.Context) {
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get current user"})
		return
	}

	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), currentUser, &req)
	if err != nil {
		switch err {
		case service.ErrUserNotAllowed:
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

// List lists users (admin only)
func (h *UserHandler) List(c *gin.Context) {
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get current user"})
		return
	}

	// Parse optional tenant_id filter
	var tenantID int64
	if tidStr := c.Query("tenant_id"); tidStr != "" {
		tid, err := strconv.ParseInt(tidStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
			return
		}
		tenantID = tid
	}

	users, err := h.userService.ListUsers(c.Request.Context(), currentUser, tenantID)
	if err != nil {
		if err == service.ErrUserNotAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Get gets a user by ID (admin only)
func (h *UserHandler) Get(c *gin.Context) {
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get current user"})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), currentUser, userID)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case service.ErrUserNotAllowed:
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update updates a user (admin only)
func (h *UserHandler) Update(c *gin.Context) {
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get current user"})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), currentUser, userID, &req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case service.ErrUserNotAllowed:
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		case service.ErrCannotModifySelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot modify your own account"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete disables a user (admin only)
func (h *UserHandler) Delete(c *gin.Context) {
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get current user"})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	err = h.userService.DisableUser(c.Request.Context(), currentUser, userID)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case service.ErrUserNotAllowed:
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		case service.ErrCannotModifySelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot disable your own account"})
		case service.ErrCannotDeleteAdmin:
			c.JSON(http.StatusForbidden, gin.H{"error": "cannot disable super admin"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user disabled"})
}

// ResetPassword resets a user's password (admin only)
func (h *UserHandler) ResetPassword(c *gin.Context) {
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get current user"})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.userService.ResetPassword(c.Request.Context(), currentUser, userID, &req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case service.ErrUserNotAllowed:
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

// GetMe gets the current user's info along with tenant
func (h *UserHandler) GetMe(c *gin.Context) {
	currentUser, err := h.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get current user"})
		return
	}

	// Also fetch tenant info
	tenant, err := h.tenantRepo.GetByID(c.Request.Context(), currentUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":   currentUser,
		"tenant": tenant,
	})
}
