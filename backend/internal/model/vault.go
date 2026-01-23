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
	IsPersonal  bool      `gorm:"default:false" json:"is_personal"` // true = personal vault (only owner can see)
	OwnerID     int64     `gorm:"index" json:"owner_id,omitempty"`  // Owner ID for personal vaults
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Tenant      *Tenant       `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Owner       *User         `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
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
	Role      string    `gorm:"size:50;not null;default:'viewer'" json:"role"` // owner, admin, editor, viewer
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Vault *Vault `gorm:"foreignKey:VaultID" json:"vault,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (VaultMember) TableName() string {
	return "vault_members"
}

// Vault role constants
const (
	VaultRoleOwner  = "owner"  // All permissions
	VaultRoleAdmin  = "admin"  // Create, delete, view credentials; manage members
	VaultRoleEditor = "editor" // Create, edit, view credentials (cannot delete)
	VaultRoleViewer = "viewer" // Read-only access
)

// CanManageMembers checks if the role can manage vault members
func CanManageMembers(role string) bool {
	return role == VaultRoleOwner || role == VaultRoleAdmin
}

// CanDeleteCredentials checks if the role can delete credentials
func CanDeleteCredentials(role string) bool {
	return role == VaultRoleOwner || role == VaultRoleAdmin
}

// CanEditCredentials checks if the role can edit credentials
func CanEditCredentials(role string) bool {
	return role == VaultRoleOwner || role == VaultRoleAdmin || role == VaultRoleEditor
}

// CanViewCredentials checks if the role can view credentials
func CanViewCredentials(role string) bool {
	return role == VaultRoleOwner || role == VaultRoleAdmin || role == VaultRoleEditor || role == VaultRoleViewer
}
