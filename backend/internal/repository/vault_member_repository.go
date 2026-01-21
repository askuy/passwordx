package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/heartalkai/passwordx/internal/model"
)

type VaultMemberRepository struct {
	db *gorm.DB
}

func NewVaultMemberRepository(db *gorm.DB) *VaultMemberRepository {
	return &VaultMemberRepository{db: db}
}

func (r *VaultMemberRepository) Create(ctx context.Context, member *model.VaultMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *VaultMemberRepository) GetByVaultAndUser(ctx context.Context, vaultID, userID int64) (*model.VaultMember, error) {
	var member model.VaultMember
	err := r.db.WithContext(ctx).
		Where("vault_id = ? AND user_id = ?", vaultID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *VaultMemberRepository) Update(ctx context.Context, member *model.VaultMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *VaultMemberRepository) Delete(ctx context.Context, vaultID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("vault_id = ? AND user_id = ?", vaultID, userID).
		Delete(&model.VaultMember{}).Error
}

func (r *VaultMemberRepository) ListByVaultID(ctx context.Context, vaultID int64) ([]model.VaultMember, error) {
	var members []model.VaultMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("vault_id = ?", vaultID).
		Find(&members).Error
	return members, err
}

func (r *VaultMemberRepository) HasAccess(ctx context.Context, vaultID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.VaultMember{}).
		Where("vault_id = ? AND user_id = ?", vaultID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *VaultMemberRepository) HasRole(ctx context.Context, vaultID, userID int64, roles []string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.VaultMember{}).
		Where("vault_id = ? AND user_id = ? AND role IN ?", vaultID, userID, roles).
		Count(&count).Error
	return count > 0, err
}
