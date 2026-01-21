package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      int64     `gorm:"index;not null" json:"tenant_id"`
	Email         string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"size:255" json:"-"`
	MasterKeySalt string    `gorm:"size:64" json:"-"`
	OAuthProvider string    `gorm:"size:50" json:"oauth_provider,omitempty"`
	OAuthID       string    `gorm:"size:255" json:"-"`
	Name          string    `gorm:"size:255" json:"name"`
	Avatar        string    `gorm:"size:500" json:"avatar,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Tenant       *Tenant       `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	VaultMembers []VaultMember `gorm:"foreignKey:UserID" json:"vault_members,omitempty"`
}

func (User) TableName() string {
	return "users"
}
