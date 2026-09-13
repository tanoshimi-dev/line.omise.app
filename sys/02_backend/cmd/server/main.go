// Command server is the line.omise.app API entrypoint.
package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	authprovider "github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/auth"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/config"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/database"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/handler"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/middleware"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
)

func main() {
	cfg := config.Load()
	if cfg.SessionSecret == "" {
		log.Fatal("SESSION_SECRET is not set")
	}
	if cfg.DatabaseURL == "" {
		// Unlike reachability (handled gracefully — see internal/database),
		// an unconfigured DATABASE_URL means auth (dev-plan-04-auth) has no
		// user/session store at all, so failing fast here is clearer than a
		// nil-pointer panic on the first /auth/* request.
		log.Fatal("DATABASE_URL is not set")
	}

	dbPool, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: failed to connect: %v", err)
	}
	defer dbPool.Close()

	userRepo := repository.NewUserRepository(dbPool.DB())
	sessionRepo := repository.NewSessionRepository(dbPool.DB())

	authHandler := &handler.AuthHandler{
		Providers: map[string]authprovider.Provider{
			"line":   authprovider.NewLineClient(cfg.LineLoginChannelID, cfg.LineLoginChannelSecret, cfg.LineLoginCallbackURL),
			"google": authprovider.NewGoogleClient(cfg.GoogleOAuthClientID, cfg.GoogleOAuthClientSecret, cfg.GoogleOAuthCallbackURL),
		},
		Users:         userRepo,
		Sessions:      sessionRepo,
		AdminEmails:   cfg.AdminEmails,
		SessionSecret: cfg.SessionSecret,
		FrontendURL:   cfg.FrontendURL,
		CookieSecure:  cfg.Env == "production",
	}
	requireReader := middleware.RequireReader(sessionRepo, userRepo, cfg.SessionSecret)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	router.GET("/health", handler.Health(dbPool))

	authGroup := router.Group("/auth")
	authGroup.GET("/line/login", authHandler.Login("line"))
	authGroup.GET("/line/callback", authHandler.Callback("line"))
	authGroup.GET("/google/login", authHandler.Login("google"))
	authGroup.GET("/google/callback", authHandler.Callback("google"))
	authGroup.POST("/logout", authHandler.Logout)
	authGroup.GET("/me", requireReader, authHandler.Me)

	log.Printf("line-api listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
