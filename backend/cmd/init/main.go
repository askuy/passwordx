package cmdinit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/askuy/passwordx/backend/cmd"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/cobra"
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

var CmdRun = &cobra.Command{
	Use:                "init",
	Short:              "init passwordx",
	Long:               `init passwordx`,
	Run:                CmdFunc,
	DisableFlagParsing: true,
}

func init() {
	cmd.RootCommand.AddCommand(CmdRun)
}

func CmdFunc(cmd *cobra.Command, args []string) {
	if err := ego.New().
		Invoker(func() error {
			initUser()
			return nil
		}).
		Run(); err != nil {
		elog.Panic("startup failed", elog.FieldErr(err))
	}
}

func initUser() error {
	// Validate required parameters
	//if email == "" {
	//	fmt.Println("Error: --email is required")
	//	fmt.Println("Usage: go run cmd/init/main.go --config=config/config.toml --email=admin@example.com --password=xxx [--name=\"Admin\"]")
	//	os.Exit(1)
	//}
	//if password == "" {
	//	fmt.Println("Error: --password is required")
	//	fmt.Println("Usage: go run cmd/init/main.go --config=config/config.toml --email=admin@example.com --password=xxx [--name=\"Admin\"]")
	//	os.Exit(1)
	//}
	email = "admin@passworx.com"
	password = "passwordx"
	if len(password) < 8 {
		fmt.Println("Error: password must be at least 8 characters")
		os.Exit(1)
	}

	// Initialize database
	db := repository.InitDB()
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Use background context for database operations
	stdCtx := context.Background()

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
