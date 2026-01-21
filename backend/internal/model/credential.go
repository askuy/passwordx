package model

import (
	"time"
)

// Credential represents an encrypted password entry
type Credential struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	VaultID           int64     `gorm:"index;not null" json:"vault_id"`
	TenantID          int64     `gorm:"index;not null" json:"tenant_id"`
	TitleEncrypted    string    `gorm:"size:500;not null" json:"title_encrypted"`
	URLEncrypted      string    `gorm:"size:2000" json:"url_encrypted,omitempty"`
	UsernameEncrypted string    `gorm:"size:500" json:"username_encrypted,omitempty"`
	PasswordEncrypted string    `gorm:"size:1000;not null" json:"password_encrypted"`
	NotesEncrypted    string    `gorm:"type:text" json:"notes_encrypted,omitempty"`
	Category          string    `gorm:"size:100" json:"category,omitempty"`
	Favicon           string    `gorm:"size:500" json:"favicon,omitempty"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Vault  *Vault  `gorm:"foreignKey:VaultID" json:"vault,omitempty"`
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

func (Credential) TableName() string {
	return "credentials"
}

// CredentialDTO is the decrypted representation sent to/from clients
type CredentialDTO struct {
	ID        int64     `json:"id"`
	VaultID   int64     `json:"vault_id"`
	Title     string    `json:"title"`
	URL       string    `json:"url,omitempty"`
	Username  string    `json:"username,omitempty"`
	Password  string    `json:"password"`
	Notes     string    `json:"notes,omitempty"`
	Category  string    `json:"category,omitempty"`
	Favicon   string    `json:"favicon,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
