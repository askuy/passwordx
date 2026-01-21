package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/heartalkai/passwordx/internal/model"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *TenantRepository) GetByID(ctx context.Context, id int64) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.WithContext(ctx).First(&tenant, id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *TenantRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Tenant{}, id).Error
}

func (r *TenantRepository) ListByUserID(ctx context.Context, userID int64) ([]model.Tenant, error) {
	var tenants []model.Tenant
	err := r.db.WithContext(ctx).
		Joins("JOIN users ON users.tenant_id = tenants.id").
		Where("users.id = ?", userID).
		Find(&tenants).Error
	return tenants, err
}
