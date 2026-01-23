package model

import (
	"time"
)

// User role constants
const (
	UserRoleSuperAdmin = "super_admin" // System-level administrator, can manage all tenants and users
	UserRoleAdmin      = "admin"       // Tenant administrator, can manage users in their tenant
	UserRoleUser       = "user"        // Regular user
)

// User account type constants
const (
	AccountTypePersonal = "personal" // Personal account (independent, not part of a team)
	AccountTypeTeam     = "team"     // Team account (belongs to a tenant/team)
)

// User status constants
const (
	UserStatusActive   = "active"   // Account is active
	UserStatusInactive = "inactive" // Account is disabled
	UserStatusInvited  = "invited"  // Account is invited but not yet activated
)

// User represents a user in the system
type User struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      int64     `gorm:"index;not null" json:"tenant_id"`
	Email         string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"size:255" json:"-"`
	MasterKeySalt string    `gorm:"size:64" json:"master_key_salt,omitempty"`
	OAuthProvider string    `gorm:"size:50" json:"oauth_provider,omitempty"`
	OAuthID       string    `gorm:"size:255" json:"-"`
	Name          string    `gorm:"size:255" json:"name"`
	Avatar        string    `gorm:"size:500" json:"avatar,omitempty"`
	Role          string    `gorm:"size:50;default:'user'" json:"role"`         // super_admin, admin, user
	AccountType   string    `gorm:"size:50;default:'team'" json:"account_type"` // personal, team
	Status        string    `gorm:"size:50;default:'active'" json:"status"`     // active, inactive, invited
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Tenant       *Tenant       `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	VaultMembers []VaultMember `gorm:"foreignKey:UserID" json:"vault_members,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// IsSuperAdmin checks if the user is a super admin
func (u *User) IsSuperAdmin() bool {
	return u.Role == UserRoleSuperAdmin
}

// IsAdmin checks if the user is an admin (super_admin or admin)
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleSuperAdmin || u.Role == UserRoleAdmin
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}
