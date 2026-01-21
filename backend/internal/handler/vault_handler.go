package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/heartalkai/passwordx/internal/middleware"
	"github.com/heartalkai/passwordx/internal/service"
)

type VaultHandler struct {
	vaultService *service.VaultService
}

func NewVaultHandler(vaultService *service.VaultService) *VaultHandler {
	return &VaultHandler{
		vaultService: vaultService,
	}
}

// Create creates a new vault
func (h *VaultHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tenantID := middleware.GetTenantID(c)

	var req service.CreateVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vault, err := h.vaultService.Create(c.Request.Context(), tenantID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, vault)
}

// Get retrieves a vault by ID
func (h *VaultHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vault ID"})
		return
	}

	vault, err := h.vaultService.Get(c.Request.Context(), id, userID)
	if err != nil {
		if err == service.ErrVaultNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "vault not found"})
			return
		}
		if err == service.ErrVaultAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vault)
}

// List returns all vaults for the current user
func (h *VaultHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tenantID := middleware.GetTenantID(c)

	vaults, err := h.vaultService.List(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"vaults": vaults})
}

// Update updates a vault
func (h *VaultHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vault ID"})
		return
	}

	var req service.UpdateVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vault, err := h.vaultService.Update(c.Request.Context(), id, userID, &req)
	if err != nil {
		if err == service.ErrVaultNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "vault not found"})
			return
		}
		if err == service.ErrVaultAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vault)
}

// Delete deletes a vault
func (h *VaultHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vault ID"})
		return
	}

	if err := h.vaultService.Delete(c.Request.Context(), id, userID); err != nil {
		if err == service.ErrVaultAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddMember adds a member to a vault
func (h *VaultHandler) AddMember(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vault ID"})
		return
	}

	var req service.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.vaultService.AddMember(c.Request.Context(), id, userID, &req)
	if err != nil {
		if err == service.ErrVaultAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// RemoveMember removes a member from a vault
func (h *VaultHandler) RemoveMember(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vault ID"})
		return
	}

	targetUserID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.vaultService.RemoveMember(c.Request.Context(), id, userID, targetUserID); err != nil {
		if err == service.ErrVaultAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
