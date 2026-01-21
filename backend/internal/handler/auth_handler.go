package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/econf"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/askuy/passwordx/backend/internal/service"
)

type AuthHandler struct {
	authService  *service.AuthService
	googleConfig *oauth2.Config
	githubConfig *oauth2.Config
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	h := &AuthHandler{
		authService: authService,
	}

	// Initialize Google OAuth config
	googleClientID := econf.GetString("oauth.google.clientId")
	googleClientSecret := econf.GetString("oauth.google.clientSecret")
	if googleClientID != "" && googleClientSecret != "" {
		h.googleConfig = &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  econf.GetString("oauth.google.redirectUrl"),
			Scopes:       []string{"email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	// Initialize GitHub OAuth config
	githubClientID := econf.GetString("oauth.github.clientId")
	githubClientSecret := econf.GetString("oauth.github.clientSecret")
	if githubClientID != "" && githubClientSecret != "" {
		h.githubConfig = &oauth2.Config{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  econf.GetString("oauth.github.redirectUrl"),
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	}

	return h
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrUserExists {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// OAuthLogin initiates OAuth flow
func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	provider := c.Param("provider")

	var config *oauth2.Config
	switch provider {
	case "google":
		config = h.googleConfig
	case "github":
		config = h.githubConfig
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported OAuth provider"})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OAuth provider not configured"})
		return
	}

	// Generate state token (in production, use a proper state management)
	state := "random-state-token"
	url := config.AuthCodeURL(state)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// OAuthCallback handles OAuth callback
func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code required"})
		return
	}

	var config *oauth2.Config
	switch provider {
	case "google":
		config = h.googleConfig
	case "github":
		config = h.githubConfig
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported OAuth provider"})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OAuth provider not configured"})
		return
	}

	// Exchange code for token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}

	// Get user info based on provider
	var oauthID, email, name, avatar string

	switch provider {
	case "google":
		oauthID, email, name, avatar, err = h.getGoogleUserInfo(token)
	case "github":
		oauthID, email, name, avatar, err = h.getGitHubUserInfo(token)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	// Login or register user
	resp, err := h.authService.OAuthLogin(c.Request.Context(), provider, oauthID, email, name, avatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redirect to frontend with token
	frontendURL := econf.GetString("app.frontendUrl")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/auth/callback?token="+resp.Token)
}

func (h *AuthHandler) getGoogleUserInfo(token *oauth2.Token) (string, string, string, string, error) {
	client := h.googleConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", "", "", err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", "", "", "", err
	}

	return userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture, nil
}

func (h *AuthHandler) getGitHubUserInfo(token *oauth2.Token) (string, string, string, string, error) {
	client := h.githubConfig.Client(context.Background(), token)

	// Get user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return "", "", "", "", err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", "", "", "", err
	}

	// Get primary email
	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", "", "", "", err
	}
	defer emailResp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
		return "", "", "", "", err
	}

	var primaryEmail string
	for _, e := range emails {
		if e.Primary && e.Verified {
			primaryEmail = e.Email
			break
		}
	}

	name := userInfo.Name
	if name == "" {
		name = userInfo.Login
	}

	return fmt.Sprintf("%d", userInfo.ID), primaryEmail, name, userInfo.AvatarURL, nil
}
