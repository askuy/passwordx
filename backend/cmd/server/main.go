package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egin"

	"github.com/askuy/passwordx/backend/internal/handler"
	"github.com/askuy/passwordx/backend/internal/middleware"
	"github.com/askuy/passwordx/backend/internal/repository"
	"github.com/askuy/passwordx/backend/internal/service"
)

func main() {
	if err := ego.New().
		Invoker(initDependencies).
		Serve(newHTTPServer()).
		Run(); err != nil {
		elog.Panic("startup failed", elog.FieldErr(err))
	}
}

var (
	authHandler       *handler.AuthHandler
	tenantHandler     *handler.TenantHandler
	vaultHandler      *handler.VaultHandler
	credentialHandler *handler.CredentialHandler
	authMiddleware    *middleware.AuthMiddleware
)

func initDependencies() error {
	// Initialize database
	db := repository.InitDB()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	vaultRepo := repository.NewVaultRepository(db)
	credentialRepo := repository.NewCredentialRepository(db)
	vaultMemberRepo := repository.NewVaultMemberRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, tenantRepo)
	tenantService := service.NewTenantService(tenantRepo, userRepo)
	vaultService := service.NewVaultService(vaultRepo, vaultMemberRepo)
	credentialService := service.NewCredentialService(credentialRepo, vaultMemberRepo)

	// Initialize handlers
	authHandler = handler.NewAuthHandler(authService)
	tenantHandler = handler.NewTenantHandler(tenantService)
	vaultHandler = handler.NewVaultHandler(vaultService)
	credentialHandler = handler.NewCredentialHandler(credentialService)

	// Initialize middleware
	authMiddleware = middleware.NewAuthMiddleware()

	return nil
}

func newHTTPServer() *egin.Component {
	server := egin.Load("server.http").Build()

	// CORS middleware
	server.Use(middleware.CORS())

	// Public routes
	api := server.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/oauth/:provider", authHandler.OAuthLogin)
			auth.GET("/oauth/:provider/callback", authHandler.OAuthCallback)
		}
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(authMiddleware.JWT())
	{
		// Tenant routes
		tenants := protected.Group("/tenants")
		{
			tenants.POST("", tenantHandler.Create)
			tenants.GET("", tenantHandler.List)
			tenants.GET("/:id", tenantHandler.Get)
			tenants.PUT("/:id", tenantHandler.Update)
			tenants.DELETE("/:id", tenantHandler.Delete)
		}

		// Vault routes
		vaults := protected.Group("/vaults")
		{
			vaults.POST("", vaultHandler.Create)
			vaults.GET("", vaultHandler.List)
			vaults.GET("/:id", vaultHandler.Get)
			vaults.PUT("/:id", vaultHandler.Update)
			vaults.DELETE("/:id", vaultHandler.Delete)
			vaults.POST("/:id/members", vaultHandler.AddMember)
			vaults.DELETE("/:id/members/:userId", vaultHandler.RemoveMember)

			// Credential routes (nested under vaults)
			vaults.POST("/:id/credentials", credentialHandler.Create)
			vaults.GET("/:id/credentials", credentialHandler.List)
			vaults.GET("/:id/credentials/:credId", credentialHandler.Get)
			vaults.PUT("/:id/credentials/:credId", credentialHandler.Update)
			vaults.DELETE("/:id/credentials/:credId", credentialHandler.Delete)
		}

		// Search credentials across all vaults
		protected.GET("/credentials/search", credentialHandler.Search)
	}

	return server
}
