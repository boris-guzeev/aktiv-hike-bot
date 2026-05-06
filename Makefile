include .env
export

seeds:
	docker run --rm \
	--network=aktiv-hike-bot_app_net \
	-e DB_DSN=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB}?sslmode=disable \
	-e TZ=$(TZ) \
	-v $(PWD):/app \
	-w /app \
	golang:1.25 \
	sh -c "go mod download && go run ./cmd/seeds"

migrate:
	docker compose run --rm migrate up

migrate-down:
	docker compose run --rm migrate down 1