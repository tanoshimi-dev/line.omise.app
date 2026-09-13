// Command server is the line.omise.app API entrypoint.
package main

import (
	"context"
	"log"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/config"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/database"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/server"
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

	router := server.New(cfg, dbPool)

	log.Printf("line-api listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
