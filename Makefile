include .env
export

.PHONY: seeds migrate migrate-down backup-db restore-db

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

restore-db:
	@set -eu; \
	if [ -z "$(BACKUP_FILE)" ]; then \
		echo "Usage: make restore-db BACKUP_FILE=backups/<file>.dump CONFIRM_RESTORE=yes"; \
		exit 1; \
	fi; \
	if [ ! -f "$(BACKUP_FILE)" ]; then \
		echo "Backup file not found: $(BACKUP_FILE)"; \
		exit 1; \
	fi; \
	if [ "$(CONFIRM_RESTORE)" != "yes" ]; then \
		echo "Restore replaces the entire $(POSTGRES_DB) database."; \
		echo "Run again with CONFIRM_RESTORE=yes to continue."; \
		exit 1; \
	fi; \
	docker compose exec -T db pg_restore --list < "$(BACKUP_FILE)" > /dev/null; \
	services_to_restart=""; \
	if [ -n "$$(docker compose ps -q --status running admin-bot)" ]; then \
		services_to_restart="$$services_to_restart admin-bot"; \
	fi; \
	if [ -n "$$(docker compose ps -q --status running client-bot)" ]; then \
		services_to_restart="$$services_to_restart client-bot"; \
	fi; \
	docker compose stop admin-bot client-bot; \
	docker compose exec -T db dropdb \
		--username="$(POSTGRES_USER)" \
		--force \
		"$(POSTGRES_DB)"; \
	docker compose exec -T db createdb \
		--username="$(POSTGRES_USER)" \
		--owner="$(POSTGRES_USER)" \
		"$(POSTGRES_DB)"; \
	if ! docker compose exec -T db pg_restore \
		--username="$(POSTGRES_USER)" \
		--dbname="$(POSTGRES_DB)" \
		--exit-on-error \
		--no-owner \
		--no-acl < "$(BACKUP_FILE)"; then \
		echo "Restore failed. Bot services remain stopped."; \
		exit 1; \
	fi; \
	if [ -n "$$services_to_restart" ]; then \
		docker compose start $$services_to_restart; \
	fi; \
	echo "Database restored from: $(BACKUP_FILE)"
