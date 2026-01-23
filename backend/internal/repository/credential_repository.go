package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/askuy/passwordx/backend/internal/model"
)

type CredentialRepository struct {
	db *gorm.DB
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

func (r *CredentialRepository) Create(ctx context.Context, credential *model.Credential) error {
	return r.db.WithContext(ctx).Create(credential).Error
}

func (r *CredentialRepository) GetByID(ctx context.Context, id int64) (*model.Credential, error) {
	var credential model.Credential
	err := r.db.WithContext(ctx).First(&credential, id).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *CredentialRepository) Update(ctx context.Context, credential *model.Credential) error {
	return r.db.WithContext(ctx).Save(credential).Error
}

func (r *CredentialRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Credential{}, id).Error
}

func (r *CredentialRepository) ListByVaultID(ctx context.Context, vaultID int64) ([]model.Credential, error) {
	var credentials []model.Credential
	err := r.db.WithContext(ctx).Where("vault_id = ?", vaultID).Find(&credentials).Error
	return credentials, err
}

func (r *CredentialRepository) ListByTenantID(ctx context.Context, tenantID int64) ([]model.Credential, error) {
	var credentials []model.Credential
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&credentials).Error
	return credentials, err
}

// SearchByURL is deprecated - searching encrypted data doesn't work
// This method now returns all credentials and filtering should be done client-side after decryption
func (r *CredentialRepository) SearchByURL(ctx context.Context, tenantID int64, userID int64, urlPattern string) ([]model.Credential, error) {
	// Since URL is encrypted, we cannot search on it server-side
	// Return all credentials and let the client filter after decryption
	return r.ListByUserVaults(ctx, tenantID, userID)
}

func (r *CredentialRepository) ListByUserVaults(ctx context.Context, tenantID int64, userID int64) ([]model.Credential, error) {
	var credentials []model.Credential
	err := r.db.WithContext(ctx).
		Joins("JOIN vault_members ON vault_members.vault_id = credentials.vault_id").
		Where("credentials.tenant_id = ? AND vault_members.user_id = ?", tenantID, userID).
		Find(&credentials).Error
	return credentials, err
}
