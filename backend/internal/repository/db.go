package repository

import (
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/askuy/passwordx/backend/internal/model"
)

var db *gorm.DB

// InitDB initializes database connection and runs migrations
func InitDB() *gorm.DB {
	dsn := econf.GetString("mysql.default.dsn")
	debug := econf.GetBool("mysql.default.debug")

	var logLevel logger.LogLevel
	if debug {
		logLevel = logger.Info
	} else {
		logLevel = logger.Silent
	}

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		elog.Panic("failed to connect database", elog.FieldErr(err))
	}

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
