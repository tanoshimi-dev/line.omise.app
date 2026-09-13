// Command migrate applies or rolls back the database schema migrations in
// sys/02_backend/migrations. See dev-plan-02-database.md.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/config"
)

func main() {
	path := flag.String("path", "migrations", "directory containing .up.sql/.down.sql files")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: migrate [-path dir] <up|down|version>")
		flag.PrintDefaults()
	}
	flag.Parse()

	cmd := flag.Arg(0)
	if cmd == "" {
		flag.Usage()
		log.Fatal("missing command")
	}

	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	m, err := migrate.New("file://"+*path, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("migrate: failed to initialize: %v", err)
	}
	defer m.Close()

	switch cmd {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "version":
		version, dirty, verr := m.Version()
		if verr != nil {
			log.Fatalf("migrate: version: %v", verr)
		}
		fmt.Printf("version=%d dirty=%v\n", version, dirty)
		return
	default:
		flag.Usage()
		log.Fatalf("unknown command %q", cmd)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate: %s failed: %v", cmd, err)
	}
	log.Printf("migrate: %s complete", cmd)
}
