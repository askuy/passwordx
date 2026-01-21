package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gotomicro/ego/core/econf"
	"gorm.io/gorm"

	"github.com/heartalkai/passwordx/internal/model"
	"github.com/heartalkai/passwordx/internal/pkg/crypto"
	"github.com/heartalkai/passwordx/internal/repository"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	jwtSecret  string
	jwtExpire  int
}

func NewAuthService(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		jwtSecret:  econf.GetString("jwt.secret"),
		jwtExpire:  econf.GetInt("jwt.expireHours"),
	}
}

type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	Name       string `json:"name" binding:"required"`
	TenantName string `json:"tenant_name" binding:"required"`
	TenantSlug string `json:"tenant_slug" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token    string        `json:"token"`
	User     *model.User   `json:"user"`
	Tenant   *model.Tenant `json:"tenant"`
	ExpireAt time.Time     `json:"expire_at"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	TenantID int64  `json:"tenant_id"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// Register creates a new user with a new tenant
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Check if tenant slug already exists
	_, err = s.tenantRepo.GetBySlug(ctx, req.TenantSlug)
	if err == nil {
		return nil, errors.New("tenant slug already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create tenant
	tenant := &model.Tenant{
		Name: req.TenantName,
		Slug: strings.ToLower(req.TenantSlug),
	}
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := crypto.HashPasswordBcrypt(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate salt for master key derivation
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		TenantID:      tenant.ID,
		Email:         req.Email,
		Name:          req.Name,
		PasswordHash:  passwordHash,
		MasterKeySalt: salt,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, expireAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:    token,
		User:     user,
		Tenant:   tenant,
		ExpireAt: expireAt,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if !crypto.VerifyPasswordBcrypt(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Get tenant
	tenant, err := s.tenantRepo.GetByID(ctx, user.TenantID)
	if err != nil {
		return nil, err
	}

	// Generate JWT token
	token, expireAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:    token,
		User:     user,
		Tenant:   tenant,
		ExpireAt: expireAt,
	}, nil
}

// OAuthLogin handles OAuth authentication
func (s *AuthService) OAuthLogin(ctx context.Context, provider, oauthID, email, name, avatar string) (*AuthResponse, error) {
	// Try to find existing user by OAuth
	user, err := s.userRepo.GetByOAuth(ctx, provider, oauthID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var tenant *model.Tenant

	if user == nil {
		// Check if user exists by email
		user, err = s.userRepo.GetByEmail(ctx, email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		if user != nil {
			// Link OAuth to existing user
			user.OAuthProvider = provider
			user.OAuthID = oauthID
			if user.Avatar == "" && avatar != "" {
				user.Avatar = avatar
			}
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, err
			}
			tenant, err = s.tenantRepo.GetByID(ctx, user.TenantID)
			if err != nil {
				return nil, err
			}
		} else {
			// Create new tenant and user
			tenantSlug := strings.ToLower(strings.ReplaceAll(name, " ", "-")) + "-" + oauthID[:8]
			tenant = &model.Tenant{
				Name: name + "'s Workspace",
				Slug: tenantSlug,
			}
			if err := s.tenantRepo.Create(ctx, tenant); err != nil {
				return nil, err
			}

			salt, err := crypto.GenerateSalt()
			if err != nil {
				return nil, err
			}

			user = &model.User{
				TenantID:      tenant.ID,
				Email:         email,
				Name:          name,
				Avatar:        avatar,
				OAuthProvider: provider,
				OAuthID:       oauthID,
				MasterKeySalt: salt,
			}
			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, err
			}
		}
	} else {
		tenant, err = s.tenantRepo.GetByID(ctx, user.TenantID)
		if err != nil {
			return nil, err
		}
	}

	// Generate JWT token
	token, expireAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:    token,
		User:     user,
		Tenant:   tenant,
		ExpireAt: expireAt,
	}, nil
}

// GetUserSalt returns the master key salt for a user
func (s *AuthService) GetUserSalt(ctx context.Context, userID int64) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return user.MasterKeySalt, nil
}

func (s *AuthService) generateToken(user *model.User) (string, time.Time, error) {
	expireAt := time.Now().Add(time.Duration(s.jwtExpire) * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expireAt, nil
}
