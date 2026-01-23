package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/askuy/passwordx/backend/internal/model"
	"github.com/askuy/passwordx/backend/internal/repository"
)

var (
	ErrVaultNotFound          = errors.New("vault not found")
	ErrVaultAccessDenied      = errors.New("vault access denied")
	ErrPersonalVaultNoMembers = errors.New("personal vaults cannot have additional members")
	ErrInvalidVaultRole       = errors.New("invalid vault role")
)

type VaultService struct {
	vaultRepo       *repository.VaultRepository
	vaultMemberRepo *repository.VaultMemberRepository
}

func NewVaultService(vaultRepo *repository.VaultRepository, vaultMemberRepo *repository.VaultMemberRepository) *VaultService {
	return &VaultService{
		vaultRepo:       vaultRepo,
		vaultMemberRepo: vaultMemberRepo,
	}
}

type CreateVaultRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IsPersonal  bool   `json:"is_personal"` // true = personal vault (only owner can see)
}

type UpdateVaultRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type AddMemberRequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin editor viewer"`
}

// Create creates a new vault and adds the creator as owner
func (s *VaultService) Create(ctx context.Context, tenantID, userID int64, req *CreateVaultRequest) (*model.Vault, error) {
	vault := &model.Vault{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		IsPersonal:  req.IsPersonal,
	}

	// For personal vaults, set the owner ID directly
	if req.IsPersonal {
		vault.OwnerID = userID
	}

	if err := s.vaultRepo.Create(ctx, vault); err != nil {
		return nil, err
	}

	// Add creator as owner (for both personal and team vaults)
	member := &model.VaultMember{
		VaultID: vault.ID,
		UserID:  userID,
		Role:    model.VaultRoleOwner,
	}
	if err := s.vaultMemberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return vault, nil
}

// Get retrieves a vault by ID with access check
func (s *VaultService) Get(ctx context.Context, vaultID, userID int64) (*model.Vault, error) {
	vault, err := s.vaultRepo.GetByIDWithMembers(ctx, vaultID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, err
	}

	// For personal vaults, only the owner can access
	if vault.IsPersonal {
		if vault.OwnerID != userID {
			return nil, ErrVaultAccessDenied
		}
		return vault, nil
	}

	// For team vaults, check membership
	hasAccess, err := s.vaultMemberRepo.HasAccess(ctx, vaultID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrVaultAccessDenied
	}

	return vault, nil
}

// List returns all vaults for a user in a tenant
func (s *VaultService) List(ctx context.Context, tenantID, userID int64) ([]model.Vault, error) {
	return s.vaultRepo.ListByUserID(ctx, userID, tenantID)
}

// Update updates a vault (only admins and owners)
func (s *VaultService) Update(ctx context.Context, vaultID, userID int64, req *UpdateVaultRequest) (*model.Vault, error) {
	// Check if user is admin or owner
	hasRole, err := s.vaultMemberRepo.HasRole(ctx, vaultID, userID, []string{model.VaultRoleOwner, model.VaultRoleAdmin})
	if err != nil {
		return nil, err
	}
	if !hasRole {
		return nil, ErrVaultAccessDenied
	}

	vault, err := s.vaultRepo.GetByID(ctx, vaultID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, err
	}

	if req.Name != "" {
		vault.Name = req.Name
	}
	if req.Description != "" {
		vault.Description = req.Description
	}
	if req.Icon != "" {
		vault.Icon = req.Icon
	}

	if err := s.vaultRepo.Update(ctx, vault); err != nil {
		return nil, err
	}

	return vault, nil
}

// Delete deletes a vault (only owners)
func (s *VaultService) Delete(ctx context.Context, vaultID, userID int64) error {
	// Check if user is owner
	hasRole, err := s.vaultMemberRepo.HasRole(ctx, vaultID, userID, []string{model.VaultRoleOwner})
	if err != nil {
		return err
	}
	if !hasRole {
		return ErrVaultAccessDenied
	}

	return s.vaultRepo.Delete(ctx, vaultID)
}

// AddMember adds a member to a vault
func (s *VaultService) AddMember(ctx context.Context, vaultID, userID int64, req *AddMemberRequest) (*model.VaultMember, error) {
	// Get vault to check if it's personal
	vault, err := s.vaultRepo.GetByID(ctx, vaultID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, err
	}

	// Personal vaults cannot have additional members
	if vault.IsPersonal {
		return nil, ErrPersonalVaultNoMembers
	}

	// Validate role
	if req.Role != model.VaultRoleAdmin && req.Role != model.VaultRoleEditor && req.Role != model.VaultRoleViewer {
		return nil, ErrInvalidVaultRole
	}

	// Check if requesting user is admin or owner
	hasRole, err := s.vaultMemberRepo.HasRole(ctx, vaultID, userID, []string{model.VaultRoleOwner, model.VaultRoleAdmin})
	if err != nil {
		return nil, err
	}
	if !hasRole {
		return nil, ErrVaultAccessDenied
	}

	// Check if member already exists
	existing, err := s.vaultMemberRepo.GetByVaultAndUser(ctx, vaultID, req.UserID)
	if err == nil && existing != nil {
		// Cannot change owner's role
		if existing.Role == model.VaultRoleOwner {
			return nil, errors.New("cannot change owner's role")
		}
		// Update role
		existing.Role = req.Role
		if err := s.vaultMemberRepo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	member := &model.VaultMember{
		VaultID: vaultID,
		UserID:  req.UserID,
		Role:    req.Role,
	}

	if err := s.vaultMemberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// RemoveMember removes a member from a vault
func (s *VaultService) RemoveMember(ctx context.Context, vaultID, requestingUserID, targetUserID int64) error {
	// Check if requesting user is admin or owner
	hasRole, err := s.vaultMemberRepo.HasRole(ctx, vaultID, requestingUserID, []string{model.VaultRoleOwner, model.VaultRoleAdmin})
	if err != nil {
		return err
	}
	if !hasRole {
		return ErrVaultAccessDenied
	}

	// Cannot remove the owner
	targetMember, err := s.vaultMemberRepo.GetByVaultAndUser(ctx, vaultID, targetUserID)
	if err != nil {
		return err
	}
	if targetMember.Role == model.VaultRoleOwner {
		return errors.New("cannot remove vault owner")
	}

	return s.vaultMemberRepo.Delete(ctx, vaultID, targetUserID)
}
