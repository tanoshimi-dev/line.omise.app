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
	courseRepo := repository.NewCourseRepository(dbPool.DB())
	articleRepo := repository.NewArticleRepository(dbPool.DB())
	usecaseRepo := repository.NewUsecaseRepository(dbPool.DB())

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
	requireAdmin := middleware.RequireAdmin(sessionRepo, userRepo, cfg.SessionSecret)

	courseHandler := &handler.CourseHandler{Courses: courseRepo}
	articleHandler := &handler.ArticleHandler{Articles: articleRepo}
	usecaseHandler := &handler.UsecaseHandler{Usecases: usecaseRepo}

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

	// Public content API (dev-plan-05-content-api) — no login required.
	api := router.Group("/api")
	api.GET("/courses", courseHandler.ListCourses)
	api.GET("/courses/:slug", courseHandler.GetCourse)
	api.GET("/courses/:slug/lessons/:lessonSlug", courseHandler.GetLesson)
	api.GET("/articles", articleHandler.ListArticles)
	api.GET("/articles/:category/:slug", articleHandler.GetArticle)
	api.GET("/usecases", usecaseHandler.ListUsecases)
	api.GET("/usecases/:slug", usecaseHandler.GetUsecase)

	// Admin content API — write access requires role=admin.
	admin := api.Group("/admin", requireAdmin)
	admin.POST("/courses", courseHandler.AdminCreateCourse)
	admin.PUT("/courses/:id", courseHandler.AdminUpdateCourse)
	admin.DELETE("/courses/:id", courseHandler.AdminDeleteCourse)
	admin.POST("/courses/:id/lessons", courseHandler.AdminCreateLesson)
	admin.PUT("/lessons/:id", courseHandler.AdminUpdateLesson)
	admin.DELETE("/lessons/:id", courseHandler.AdminDeleteLesson)
	admin.POST("/articles", articleHandler.AdminCreateArticle)
	admin.PUT("/articles/:id", articleHandler.AdminUpdateArticle)
	admin.DELETE("/articles/:id", articleHandler.AdminDeleteArticle)
	admin.POST("/articles/:id/tags", articleHandler.AdminAttachTag)
	admin.POST("/usecases", usecaseHandler.AdminCreateUsecase)
	admin.PUT("/usecases/:id", usecaseHandler.AdminUpdateUsecase)
	admin.DELETE("/usecases/:id", usecaseHandler.AdminDeleteUsecase)

	log.Printf("line-api listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
