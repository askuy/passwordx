package service

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/askuy/passwordx/backend/internal/model"
	"github.com/askuy/passwordx/backend/internal/pkg/crypto"
	"github.com/askuy/passwordx/backend/internal/repository"
)

var (
	ErrUserNotAllowed    = errors.New("user not allowed to perform this action")
	ErrCannotModifySelf  = errors.New("cannot modify your own account")
	ErrCannotDeleteAdmin = errors.New("cannot delete super admin")
)

type UserService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
}

func NewUserService(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// CreateUserRequest is the request body for creating a user
type CreateUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Name        string `json:"name" binding:"required"`
	Password    string `json:"password"`                        // Optional, if empty user must use OAuth
	AccountType string `json:"account_type" binding:"required"` // personal, team
	TenantID    int64  `json:"tenant_id"`                       // Required for team accounts
	Role        string `json:"role"`                            // admin, user (default: user)
}

// UpdateUserRequest is the request body for updating a user
type UpdateUserRequest struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	AccountType string `json:"account_type"`
}

// ResetPasswordRequest is the request body for resetting a user's password
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

// CreateUser creates a new user (admin only)
func (s *UserService) CreateUser(ctx context.Context, currentUser *model.User, req *CreateUserRequest) (*model.User, error) {
	// Check permissions
	if !currentUser.IsAdmin() {
		return nil, ErrUserNotAllowed
	}

	// Validate account type
	if req.AccountType != model.AccountTypePersonal && req.AccountType != model.AccountTypeTeam {
		return nil, errors.New("invalid account type")
	}

	// For team accounts, tenant ID is required
	var tenantID int64
	if req.AccountType == model.AccountTypeTeam {
		if req.TenantID == 0 {
			return nil, errors.New("tenant_id is required for team accounts")
		}
		// Non-super admins can only create users in their own tenant
		if !currentUser.IsSuperAdmin() && req.TenantID != currentUser.TenantID {
			return nil, ErrUserNotAllowed
		}
		tenantID = req.TenantID
	} else {
		// Personal accounts: create a personal tenant for them
		tenant := &model.Tenant{
			Name: req.Name + "'s Personal Space",
			Slug: strings.ToLower(strings.ReplaceAll(req.Email, "@", "-at-")) + "-personal",
		}
		if err := s.tenantRepo.Create(ctx, tenant); err != nil {
			return nil, err
		}
		tenantID = tenant.ID
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Determine user status
	status := model.UserStatusInvited
	var passwordHash string

	// If password is provided, hash it and set status to active
	if req.Password != "" {
		if len(req.Password) < 8 {
			return nil, errors.New("password must be at least 8 characters")
		}
		passwordHash, err = crypto.HashPasswordBcrypt(req.Password)
		if err != nil {
			return nil, err
		}
		status = model.UserStatusActive
	}

	// Generate salt for master key derivation
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, err
	}

	// Determine role (non-super admins cannot create super admins)
	role := model.UserRoleUser
	if req.Role != "" {
		if req.Role == model.UserRoleSuperAdmin && !currentUser.IsSuperAdmin() {
			return nil, errors.New("only super admin can create super admin users")
		}
		if req.Role != model.UserRoleSuperAdmin && req.Role != model.UserRoleAdmin && req.Role != model.UserRoleUser {
			return nil, errors.New("invalid role")
		}
		role = req.Role
	}

	user := &model.User{
		TenantID:      tenantID,
		Email:         strings.ToLower(req.Email),
		Name:          req.Name,
		PasswordHash:  passwordHash,
		MasterKeySalt: salt,
		Role:          role,
		AccountType:   req.AccountType,
		Status:        status,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ListUsers lists users based on current user's permissions
func (s *UserService) ListUsers(ctx context.Context, currentUser *model.User, tenantID int64) ([]model.User, error) {
	// Super admin can see all users or filter by tenant
	if currentUser.IsSuperAdmin() {
		if tenantID > 0 {
			return s.userRepo.ListByTenantID(ctx, tenantID)
		}
		return s.userRepo.ListAll(ctx)
	}

	// Regular admins can only see users in their tenant
	if currentUser.IsAdmin() {
		return s.userRepo.ListByTenantID(ctx, currentUser.TenantID)
	}

	return nil, ErrUserNotAllowed
}

// GetUser gets a user by ID
func (s *UserService) GetUser(ctx context.Context, currentUser *model.User, userID int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Check permissions
	if !currentUser.IsSuperAdmin() && user.TenantID != currentUser.TenantID {
		return nil, ErrUserNotAllowed
	}

	return user, nil
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(ctx context.Context, currentUser *model.User, userID int64, req *UpdateUserRequest) (*model.User, error) {
	// Check permissions
	if !currentUser.IsAdmin() {
		return nil, ErrUserNotAllowed
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Non-super admins can only update users in their tenant
	if !currentUser.IsSuperAdmin() && user.TenantID != currentUser.TenantID {
		return nil, ErrUserNotAllowed
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}

	if req.Role != "" {
		// Cannot change own role
		if userID == currentUser.ID {
			return nil, ErrCannotModifySelf
		}
		// Only super admin can set super admin role
		if req.Role == model.UserRoleSuperAdmin && !currentUser.IsSuperAdmin() {
			return nil, errors.New("only super admin can assign super admin role")
		}
		// Cannot demote super admin unless you're a super admin
		if user.Role == model.UserRoleSuperAdmin && !currentUser.IsSuperAdmin() {
			return nil, errors.New("cannot modify super admin")
		}
		user.Role = req.Role
	}

	if req.Status != "" {
		// Cannot change own status
		if userID == currentUser.ID {
			return nil, ErrCannotModifySelf
		}
		if req.Status != model.UserStatusActive && req.Status != model.UserStatusInactive && req.Status != model.UserStatusInvited {
			return nil, errors.New("invalid status")
		}
		user.Status = req.Status
	}

	if req.AccountType != "" {
		if req.AccountType != model.AccountTypePersonal && req.AccountType != model.AccountTypeTeam {
			return nil, errors.New("invalid account type")
		}
		user.AccountType = req.AccountType
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DisableUser disables a user account
func (s *UserService) DisableUser(ctx context.Context, currentUser *model.User, userID int64) error {
	// Check permissions
	if !currentUser.IsAdmin() {
		return ErrUserNotAllowed
	}

	// Cannot disable yourself
	if userID == currentUser.ID {
		return ErrCannotModifySelf
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Non-super admins can only disable users in their tenant
	if !currentUser.IsSuperAdmin() && user.TenantID != currentUser.TenantID {
		return ErrUserNotAllowed
	}

	// Cannot disable super admin unless you're a super admin
	if user.Role == model.UserRoleSuperAdmin && !currentUser.IsSuperAdmin() {
		return ErrCannotDeleteAdmin
	}

	return s.userRepo.UpdateStatus(ctx, userID, model.UserStatusInactive)
}

// ResetPassword resets a user's password
func (s *UserService) ResetPassword(ctx context.Context, currentUser *model.User, userID int64, req *ResetPasswordRequest) error {
	// Check permissions
	if !currentUser.IsAdmin() {
		return ErrUserNotAllowed
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Non-super admins can only reset passwords for users in their tenant
	if !currentUser.IsSuperAdmin() && user.TenantID != currentUser.TenantID {
		return ErrUserNotAllowed
	}

	// Hash new password
	passwordHash, err := crypto.HashPasswordBcrypt(req.Password)
	if err != nil {
		return err
	}

	user.PasswordHash = passwordHash
	// If user was invited, activate them
	if user.Status == model.UserStatusInvited {
		user.Status = model.UserStatusActive
	}

	return s.userRepo.Update(ctx, user)
}
