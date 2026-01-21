package service

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/askuy/passwordx/backend/internal/model"
	"github.com/askuy/passwordx/backend/internal/repository"
)

var (
	ErrTenantNotFound  = errors.New("tenant not found")
	ErrTenantSlugTaken = errors.New("tenant slug already taken")
)

type TenantService struct {
	tenantRepo *repository.TenantRepository
	userRepo   *repository.UserRepository
}

func NewTenantService(tenantRepo *repository.TenantRepository, userRepo *repository.UserRepository) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
	}
}

type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type UpdateTenantRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Create creates a new tenant
func (s *TenantService) Create(ctx context.Context, userID int64, req *CreateTenantRequest) (*model.Tenant, error) {
	// Check if slug is taken
	_, err := s.tenantRepo.GetBySlug(ctx, strings.ToLower(req.Slug))
	if err == nil {
		return nil, ErrTenantSlugTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	tenant := &model.Tenant{
		Name: req.Name,
		Slug: strings.ToLower(req.Slug),
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	// Add user to the new tenant
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update user's tenant (or create a tenant membership system)
	user.TenantID = tenant.ID
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return tenant, nil
}

// Get retrieves a tenant by ID
func (s *TenantService) Get(ctx context.Context, id int64) (*model.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return tenant, nil
}

// List returns all tenants for a user
func (s *TenantService) List(ctx context.Context, userID int64) ([]model.Tenant, error) {
	return s.tenantRepo.ListByUserID(ctx, userID)
}

// Update updates a tenant
func (s *TenantService) Update(ctx context.Context, id int64, req *UpdateTenantRequest) (*model.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	if req.Name != "" {
		tenant.Name = req.Name
	}

	if req.Slug != "" {
		// Check if new slug is taken
		existing, err := s.tenantRepo.GetBySlug(ctx, strings.ToLower(req.Slug))
		if err == nil && existing.ID != id {
			return nil, ErrTenantSlugTaken
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		tenant.Slug = strings.ToLower(req.Slug)
	}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

// Delete deletes a tenant
func (s *TenantService) Delete(ctx context.Context, id int64) error {
	return s.tenantRepo.Delete(ctx, id)
}
