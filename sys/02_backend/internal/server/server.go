// Package server builds the Gin router. Extracted from cmd/server/main.go
// (dev-plan-12-test-phase1 12.1) so handler tests can build the exact same
// router — including the real auth middleware chain — against a test
// database, instead of duplicating route registration or mocking
// repositories behind new interfaces.
package server

import (
	"github.com/gin-gonic/gin"

	authprovider "github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/auth"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/config"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/database"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/handler"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/middleware"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
)

// New builds the full application router.
func New(cfg config.Config, dbPool *database.Pool) *gin.Engine {
	userRepo := repository.NewUserRepository(dbPool.DB())
	sessionRepo := repository.NewSessionRepository(dbPool.DB())
	articleRepo := repository.NewArticleRepository(dbPool.DB())
	usecaseRepo := repository.NewUsecaseRepository(dbPool.DB())
	quizRepo := repository.NewQuizRepository(dbPool.DB())

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
	optionalUser := middleware.OptionalUser(sessionRepo, userRepo, cfg.SessionSecret)

	articleHandler := &handler.ArticleHandler{Articles: articleRepo}
	usecaseHandler := &handler.UsecaseHandler{Usecases: usecaseRepo}
	adminQuizHandler := &handler.AdminQuizHandler{Quizzes: quizRepo}
	quizHandler := &handler.QuizHandler{Quizzes: quizRepo}

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
	authGroup.DELETE("/me", requireReader, authHandler.DeleteMe)

	// Public content API (dev-plan-05-content-api) — no login required.
	api := router.Group("/api")
	api.GET("/articles", articleHandler.ListArticles)
	api.GET("/articles/:category/:slug", articleHandler.GetArticle)
	api.GET("/usecases", usecaseHandler.ListUsecases)
	api.GET("/usecases/:slug", usecaseHandler.GetUsecase)
	api.GET("/quizzes", quizHandler.ListQuizzes)
	api.GET("/quizzes/:slug", quizHandler.GetQuiz)

	// Admin content API — write access requires role=admin. GET routes here
	// (dev-plan-11-frontend-admin) return drafts too, unlike the public GETs
	// above, so they must stay behind requireAdmin.
	admin := api.Group("/admin", requireAdmin)
	admin.GET("/articles", articleHandler.AdminListArticles)
	admin.GET("/articles/:id", articleHandler.AdminGetArticle)
	admin.POST("/articles", articleHandler.AdminCreateArticle)
	admin.PUT("/articles/:id", articleHandler.AdminUpdateArticle)
	admin.DELETE("/articles/:id", articleHandler.AdminDeleteArticle)
	admin.POST("/articles/:id/tags", articleHandler.AdminAttachTag)
	admin.GET("/usecases", usecaseHandler.AdminListUsecases)
	admin.GET("/usecases/:id", usecaseHandler.AdminGetUsecase)
	admin.POST("/usecases", usecaseHandler.AdminCreateUsecase)
	admin.PUT("/usecases/:id", usecaseHandler.AdminUpdateUsecase)
	admin.DELETE("/usecases/:id", usecaseHandler.AdminDeleteUsecase)
	admin.GET("/quizzes", adminQuizHandler.AdminListQuizzes)
	admin.GET("/quizzes/:id", adminQuizHandler.AdminGetQuiz)
	admin.POST("/quizzes", adminQuizHandler.AdminCreateQuiz)
	admin.PUT("/quizzes/:id", adminQuizHandler.AdminUpdateQuiz)
	admin.DELETE("/quizzes/:id", adminQuizHandler.AdminDeleteQuiz)
	admin.POST("/quizzes/:quizId/questions", adminQuizHandler.AdminCreateQuestion)
	admin.PUT("/quiz-questions/:id", adminQuizHandler.AdminUpdateQuestion)
	admin.DELETE("/quiz-questions/:id", adminQuizHandler.AdminDeleteQuestion)

	// Quiz/exam answer+submit — anyone may answer/submit (dev-plan-2-3
	// 2-3.2/2-3.3); optionalUser attaches the caller's identity when logged
	// in so the handler can decide whether to persist the answer/attempt.
	api.POST("/quiz-questions/:id/answer", optionalUser, quizHandler.AnswerQuestion)
	api.POST("/quizzes/:slug/submit", optionalUser, quizHandler.SubmitQuiz)

	// Quiz/exam history+progress — Reader login required (dev-plan-2-3 2-3.4).
	api.GET("/me/quizzes/progress", requireReader, quizHandler.GetMyQuizzesProgress)
	api.GET("/me/quizzes/:slug/history", requireReader, quizHandler.GetPracticeHistory)
	api.GET("/me/quizzes/:slug/attempts", requireReader, quizHandler.ListAttempts)
	api.GET("/me/quizzes/:slug/attempts/:attemptId", requireReader, quizHandler.GetAttempt)
	api.GET("/me/quizzes/:slug/progress", requireReader, quizHandler.GetQuizProgress)

	// Reader-owned history deletion (dev-plan-quiz-history-delete).
	api.DELETE("/me/quizzes/:slug/history/:answerId", requireReader, quizHandler.DeletePracticeAnswer)
	api.DELETE("/me/quizzes/:slug/history", requireReader, quizHandler.DeleteAllPracticeHistory)
	api.DELETE("/me/quizzes/:slug/attempts/:attemptId", requireReader, quizHandler.DeleteAttempt)
	api.DELETE("/me/quizzes/:slug/attempts", requireReader, quizHandler.DeleteAllAttempts)

	return router
}
