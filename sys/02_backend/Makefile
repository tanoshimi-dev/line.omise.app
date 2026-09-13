COMPOSE := docker compose -f ../01_infra/docker-compose.yml
POSTGRES_USER := line_omise
POSTGRES_DB := line_omise_app

.PHONY: migrate-up migrate-down migrate-version migrate-seed

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-version:
	go run ./cmd/migrate version

migrate-seed:
	$(COMPOSE) exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) < migrations/seed.sql
