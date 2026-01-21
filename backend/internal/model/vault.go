package model

import (
	"time"
)

// Vault represents a password vault that can contain multiple credentials
type Vault struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    int64     `gorm:"index;not null" json:"tenant_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"size:1000" json:"description,omitempty"`
	Icon        string    `gorm:"size:100" json:"icon,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Tenant      *Tenant       `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Members     []VaultMember `gorm:"foreignKey:VaultID" json:"members,omitempty"`
	Credentials []Credential  `gorm:"foreignKey:VaultID" json:"credentials,omitempty"`
}

func (Vault) TableName() string {
	return "vaults"
}

// VaultMember represents the membership of a user in a vault
type VaultMember struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	VaultID   int64     `gorm:"index;not null" json:"vault_id"`
	UserID    int64     `gorm:"index;not null" json:"user_id"`
	Role      string    `gorm:"size:50;not null;default:'member'" json:"role"` // owner, admin, member
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Vault *Vault `gorm:"foreignKey:VaultID" json:"vault,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (VaultMember) TableName() string {
	return "vault_members"
}

const (
	VaultRoleOwner  = "owner"
	VaultRoleAdmin  = "admin"
	VaultRoleMember = "member"
)
