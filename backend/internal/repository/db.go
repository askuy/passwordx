package repository

import (
	"github.com/gotomicro/ego-component/egorm"
	"github.com/gotomicro/ego/core/elog"
	"gorm.io/gorm"

	"github.com/heartalkai/passwordx/internal/model"
)

var db *gorm.DB

// InitDB initializes database connection and runs migrations
func InitDB() *gorm.DB {
	db = egorm.Load("mysql.default").Build()

	// Auto migrate models
	if err := db.AutoMigrate(
		&model.Tenant{},
		&model.User{},
		&model.Vault{},
		&model.VaultMember{},
		&model.Credential{},
	); err != nil {
		elog.Panic("failed to migrate database", elog.FieldErr(err))
	}

	elog.Info("database initialized and migrated")
	return db
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}
