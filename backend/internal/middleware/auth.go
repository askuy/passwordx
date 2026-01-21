package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gotomicro/ego/core/econf"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: econf.GetString("jwt.secret"),
	}
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	TenantID int64  `json:"tenant_id"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// JWT returns a JWT authentication middleware
func (m *AuthMiddleware) JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("email", claims.Email)

		// Allow tenant override via header (for multi-tenant switching)
		if tenantHeader := c.GetHeader("X-Tenant-ID"); tenantHeader != "" {
			// TODO: Verify user has access to this tenant
			// For now, we trust the header
		}

		c.Next()
	}
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) int64 {
	if v, exists := c.Get("user_id"); exists {
		return v.(int64)
	}
	return 0
}

// GetTenantID extracts tenant ID from gin context
func GetTenantID(c *gin.Context) int64 {
	if v, exists := c.Get("tenant_id"); exists {
		return v.(int64)
	}
	return 0
}

// GetEmail extracts email from gin context
func GetEmail(c *gin.Context) string {
	if v, exists := c.Get("email"); exists {
		return v.(string)
	}
	return ""
}
