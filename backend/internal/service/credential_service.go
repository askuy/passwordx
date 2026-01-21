package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/heartalkai/passwordx/internal/model"
	"github.com/heartalkai/passwordx/internal/repository"
)

var (
	ErrCredentialNotFound     = errors.New("credential not found")
	ErrCredentialAccessDenied = errors.New("credential access denied")
)

type CredentialService struct {
	credentialRepo  *repository.CredentialRepository
	vaultMemberRepo *repository.VaultMemberRepository
}

func NewCredentialService(credentialRepo *repository.CredentialRepository, vaultMemberRepo *repository.VaultMemberRepository) *CredentialService {
	return &CredentialService{
		credentialRepo:  credentialRepo,
		vaultMemberRepo: vaultMemberRepo,
	}
}

type CreateCredentialRequest struct {
	TitleEncrypted    string `json:"title_encrypted" binding:"required"`
	URLEncrypted      string `json:"url_encrypted"`
	UsernameEncrypted string `json:"username_encrypted"`
	PasswordEncrypted string `json:"password_encrypted" binding:"required"`
	NotesEncrypted    string `json:"notes_encrypted"`
	Category          string `json:"category"`
	Favicon           string `json:"favicon"`
}

type UpdateCredentialRequest struct {
	TitleEncrypted    string `json:"title_encrypted"`
	URLEncrypted      string `json:"url_encrypted"`
	UsernameEncrypted string `json:"username_encrypted"`
	PasswordEncrypted string `json:"password_encrypted"`
	NotesEncrypted    string `json:"notes_encrypted"`
	Category          string `json:"category"`
	Favicon           string `json:"favicon"`
}

// Create creates a new credential in a vault
func (s *CredentialService) Create(ctx context.Context, vaultID, tenantID, userID int64, req *CreateCredentialRequest) (*model.Credential, error) {
	// Check vault access
	hasAccess, err := s.vaultMemberRepo.HasAccess(ctx, vaultID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrCredentialAccessDenied
	}

	credential := &model.Credential{
		VaultID:           vaultID,
		TenantID:          tenantID,
		TitleEncrypted:    req.TitleEncrypted,
		URLEncrypted:      req.URLEncrypted,
		UsernameEncrypted: req.UsernameEncrypted,
		PasswordEncrypted: req.PasswordEncrypted,
		NotesEncrypted:    req.NotesEncrypted,
		Category:          req.Category,
		Favicon:           req.Favicon,
	}

	if err := s.credentialRepo.Create(ctx, credential); err != nil {
		return nil, err
	}

	return credential, nil
}

// Get retrieves a credential by ID with access check
func (s *CredentialService) Get(ctx context.Context, credentialID, userID int64) (*model.Credential, error) {
	credential, err := s.credentialRepo.GetByID(ctx, credentialID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}

	// Check vault access
	hasAccess, err := s.vaultMemberRepo.HasAccess(ctx, credential.VaultID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrCredentialAccessDenied
	}

	return credential, nil
}

// List returns all credentials in a vault
func (s *CredentialService) List(ctx context.Context, vaultID, userID int64) ([]model.Credential, error) {
	// Check vault access
	hasAccess, err := s.vaultMemberRepo.HasAccess(ctx, vaultID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrCredentialAccessDenied
	}

	return s.credentialRepo.ListByVaultID(ctx, vaultID)
}

// Update updates a credential
func (s *CredentialService) Update(ctx context.Context, credentialID, userID int64, req *UpdateCredentialRequest) (*model.Credential, error) {
	credential, err := s.credentialRepo.GetByID(ctx, credentialID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}

	// Check vault access
	hasAccess, err := s.vaultMemberRepo.HasAccess(ctx, credential.VaultID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrCredentialAccessDenied
	}

	if req.TitleEncrypted != "" {
		credential.TitleEncrypted = req.TitleEncrypted
	}
	if req.URLEncrypted != "" {
		credential.URLEncrypted = req.URLEncrypted
	}
	if req.UsernameEncrypted != "" {
		credential.UsernameEncrypted = req.UsernameEncrypted
	}
	if req.PasswordEncrypted != "" {
		credential.PasswordEncrypted = req.PasswordEncrypted
	}
	if req.NotesEncrypted != "" {
		credential.NotesEncrypted = req.NotesEncrypted
	}
	if req.Category != "" {
		credential.Category = req.Category
	}
	if req.Favicon != "" {
		credential.Favicon = req.Favicon
	}

	if err := s.credentialRepo.Update(ctx, credential); err != nil {
		return nil, err
	}

	return credential, nil
}

// Delete deletes a credential
func (s *CredentialService) Delete(ctx context.Context, credentialID, userID int64) error {
	credential, err := s.credentialRepo.GetByID(ctx, credentialID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCredentialNotFound
		}
		return err
	}

	// Check vault access (only admin and owner can delete)
	hasRole, err := s.vaultMemberRepo.HasRole(ctx, credential.VaultID, userID, []string{model.VaultRoleOwner, model.VaultRoleAdmin})
	if err != nil {
		return err
	}
	if !hasRole {
		return ErrCredentialAccessDenied
	}

	return s.credentialRepo.Delete(ctx, credentialID)
}

// Search searches credentials across user's vaults
func (s *CredentialService) Search(ctx context.Context, tenantID, userID int64, query string) ([]model.Credential, error) {
	if query == "" {
		return s.credentialRepo.ListByUserVaults(ctx, tenantID, userID)
	}
	return s.credentialRepo.SearchByURL(ctx, tenantID, userID, query)
}
