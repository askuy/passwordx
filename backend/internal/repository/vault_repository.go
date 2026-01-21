package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/askuy/passwordx/backend/internal/model"
)

type VaultRepository struct {
	db *gorm.DB
}

func NewVaultRepository(db *gorm.DB) *VaultRepository {
	return &VaultRepository{db: db}
}

func (r *VaultRepository) Create(ctx context.Context, vault *model.Vault) error {
	return r.db.WithContext(ctx).Create(vault).Error
}

func (r *VaultRepository) GetByID(ctx context.Context, id int64) (*model.Vault, error) {
	var vault model.Vault
	err := r.db.WithContext(ctx).First(&vault, id).Error
	if err != nil {
		return nil, err
	}
	return &vault, nil
}

func (r *VaultRepository) GetByIDWithMembers(ctx context.Context, id int64) (*model.Vault, error) {
	var vault model.Vault
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Members.User").
		First(&vault, id).Error
	if err != nil {
		return nil, err
	}
	return &vault, nil
}

func (r *VaultRepository) Update(ctx context.Context, vault *model.Vault) error {
	return r.db.WithContext(ctx).Save(vault).Error
}

func (r *VaultRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Vault{}, id).Error
}

func (r *VaultRepository) ListByTenantID(ctx context.Context, tenantID int64) ([]model.Vault, error) {
	var vaults []model.Vault
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&vaults).Error
	return vaults, err
}

func (r *VaultRepository) ListByUserID(ctx context.Context, userID int64, tenantID int64) ([]model.Vault, error) {
	var vaults []model.Vault
	err := r.db.WithContext(ctx).
		Joins("JOIN vault_members ON vault_members.vault_id = vaults.id").
		Where("vault_members.user_id = ? AND vaults.tenant_id = ?", userID, tenantID).
		Find(&vaults).Error
	return vaults, err
}
