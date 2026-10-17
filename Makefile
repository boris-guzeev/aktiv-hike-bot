include .env
export

.PHONY: seeds migrate migrate-down backup-db

BACKUP_DIR ?= backups

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

backup-db:
	@mkdir -p "$(BACKUP_DIR)"
	@set -eu; \
	backup_file="$(BACKUP_DIR)/$(POSTGRES_DB)_$$(date +%Y%m%d_%H%M%S).dump"; \
	tmp_file="$${backup_file}.tmp"; \
	trap 'rm -f "$$tmp_file"' EXIT; \
	docker compose exec -T db pg_dump \
		--username="$(POSTGRES_USER)" \
		--dbname="$(POSTGRES_DB)" \
		--format=custom \
		--no-owner \
		--no-acl > "$$tmp_file"; \
	test -s "$$tmp_file"; \
	docker compose exec -T db pg_restore --list < "$$tmp_file" > /dev/null; \
	mv "$$tmp_file" "$$backup_file"; \
	trap - EXIT; \
	echo "Database backup created: $$backup_file"
