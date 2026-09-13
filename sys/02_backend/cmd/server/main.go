// Command server is the line.omise.app API entrypoint.
package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/config"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/handler"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/middleware"
)

func main() {
	cfg := config.Load()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	router.GET("/health", handler.Health(cfg.DatabaseURL))

	log.Printf("line-api listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
