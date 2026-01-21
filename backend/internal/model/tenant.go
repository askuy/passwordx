package model

import (
	"time"
)

// Tenant represents an organization/team in the multi-tenant system
type Tenant struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Slug      string    `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Users  []User  `gorm:"foreignKey:TenantID" json:"users,omitempty"`
	Vaults []Vault `gorm:"foreignKey:TenantID" json:"vaults,omitempty"`
}

func (Tenant) TableName() string {
	return "tenants"
}
