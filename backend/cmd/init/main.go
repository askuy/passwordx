package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ejob"
	"gorm.io/gorm"

	"github.com/askuy/passwordx/backend/internal/model"
	"github.com/askuy/passwordx/backend/internal/pkg/crypto"
	"github.com/askuy/passwordx/backend/internal/repository"
)

var (
	email    string
	password string
	name     string
)

func init() {
	flag.StringVar(&email, "email", "", "Admin email address (required)")
	flag.StringVar(&password, "password", "", "Admin password (required, min 8 characters)")
	flag.StringVar(&name, "name", "Super Admin", "Admin display name")
}

func main() {
	flag.Parse()

	if err := ego.New().Job(ejob.Job("init-user", initUser)).Run(); err != nil {
		elog.Panic("init user failed", elog.FieldErr(err))
	}
}

func initUser(ctx ejob.Context) error {
	// Validate required parameters
	if email == "" {
		fmt.Println("Error: --email is required")
		fmt.Println("Usage: go run cmd/init/main.go --config=config/config.toml --email=admin@example.com --password=xxx [--name=\"Admin\"]")
		os.Exit(1)
	}
	if password == "" {
		fmt.Println("Error: --password is required")
		fmt.Println("Usage: go run cmd/init/main.go --config=config/config.toml --email=admin@example.com --password=xxx [--name=\"Admin\"]")
		os.Exit(1)
	}
	if len(password) < 8 {
		fmt.Println("Error: password must be at least 8 characters")
		os.Exit(1)
	}

	// Initialize database
	db := repository.InitDB()
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Use background context for database operations
	stdCtx := ctx.Ctx

	// Check if super admin already exists
	existingUsers, err := userRepo.ListByRole(stdCtx, model.UserRoleSuperAdmin)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing super admin: %w", err)
	}
	if len(existingUsers) > 0 {
		fmt.Printf("Super admin already exists: %s\n", existingUsers[0].Email)
		fmt.Println("If you want to create another super admin, please use the admin API.")
		return nil
	}

	// Create or get system tenant
	tenant, err := tenantRepo.GetBySlug(stdCtx, "system")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create system tenant
			tenant = &model.Tenant{
				Name: "System",
				Slug: "system",
			}
			if err := tenantRepo.Create(stdCtx, tenant); err != nil {
				return fmt.Errorf("failed to create system tenant: %w", err)
			}
			elog.Info("created system tenant")
		} else {
			return fmt.Errorf("failed to get system tenant: %w", err)
		}
	}

	// Hash password
	passwordHash, err := crypto.HashPasswordBcrypt(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate salt for master key derivation
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Create super admin user
	user := &model.User{
		TenantID:      tenant.ID,
		Email:         strings.ToLower(email),
		Name:          name,
		PasswordHash:  passwordHash,
		MasterKeySalt: salt,
		Role:          model.UserRoleSuperAdmin,
		AccountType:   model.AccountTypeTeam,
		Status:        model.UserStatusActive,
	}

	if err := userRepo.Create(stdCtx, user); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("user with email %s already exists", email)
		}
		return fmt.Errorf("failed to create super admin: %w", err)
	}

	fmt.Println("========================================")
	fmt.Println("Super admin created successfully!")
	fmt.Printf("  Email:    %s\n", user.Email)
	fmt.Printf("  Name:     %s\n", user.Name)
	fmt.Printf("  Tenant:   %s (ID: %d)\n", tenant.Name, tenant.ID)
	fmt.Println("========================================")
	fmt.Println("You can now login with these credentials.")

	return nil
}
