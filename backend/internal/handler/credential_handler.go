package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/askuy/passwordx/backend/internal/middleware"
	"github.com/askuy/passwordx/backend/internal/service"
)

type CredentialHandler struct {
	credentialService *service.CredentialService
}

func NewCredentialHandler(credentialService *service.CredentialService) *CredentialHandler {
	return &CredentialHandler{
		credentialService: credentialService,
	}
}

// Create creates a new credential
func (h *CredentialHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tenantID := middleware.GetTenantID(c)

	vaultID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vault ID"})
		return
	}

	var req service.CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credential, err := h.credentialService.Create(c.Request.Context(), vaultID, tenantID, userID, &req)
	if err != nil {
		if err == service.ErrCredentialAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, credential)
}

// Get retrieves a credential by ID
func (h *CredentialHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)

	credID, err := strconv.ParseInt(c.Param("credId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	credential, err := h.credentialService.Get(c.Request.Context(), credID, userID)
	if err != nil {
		if err == service.ErrCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		if err == service.ErrCredentialAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credential)
}

// List returns all credentials in a vault
func (h *CredentialHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	vaultID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vault ID"})
		return
	}

	credentials, err := h.credentialService.List(c.Request.Context(), vaultID, userID)
	if err != nil {
		if err == service.ErrCredentialAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credentials": credentials})
}

// Update updates a credential
func (h *CredentialHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)

	credID, err := strconv.ParseInt(c.Param("credId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	var req service.UpdateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credential, err := h.credentialService.Update(c.Request.Context(), credID, userID, &req)
	if err != nil {
		if err == service.ErrCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		if err == service.ErrCredentialAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credential)
}

// Delete deletes a credential
func (h *CredentialHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	credID, err := strconv.ParseInt(c.Param("credId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	if err := h.credentialService.Delete(c.Request.Context(), credID, userID); err != nil {
		if err == service.ErrCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		if err == service.ErrCredentialAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Search searches credentials across all user's vaults
func (h *CredentialHandler) Search(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tenantID := middleware.GetTenantID(c)
	query := c.Query("q")

	credentials, err := h.credentialService.Search(c.Request.Context(), tenantID, userID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credentials": credentials})
}
