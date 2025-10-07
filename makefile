# Run database migrations using Goose
migrate:
	@set -a; . ./.env; set +a; goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	@set -a; . ./.env; set +a; goose -dir migrations postgres "$$DATABASE_URL" down

migrate-redo:
	@set -a; . ./.env; set +a; goose -dir migrations postgres "$$DATABASE_URL" redo

migrate-create:
	@read -p "Enter your Migration name for SQL: " name; \
	goose -dir migrations create $$name sql

migrate-reset:
	@goose --version; \
	set -a; . ./.env; set +a; \
	echo "⚠️ Dropping all tables and ENUM types..."; \
	psql "$$DATABASE_URL" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
	psql "$$DATABASE_URL" -c "DROP TYPE IF EXISTS oauth_provider CASCADE;"; \
	psql "$$DATABASE_URL" -c "DROP TYPE IF EXISTS diploma_grade CASCADE;"; \
	echo "✅ Tables and ENUM types dropped. Running migrations..."; \
	goose -dir migrations postgres "$$DATABASE_URL" up; \
	echo "✅ Database reset and migrated successfully."

# =========================
# Docker Builds
# =========================
docker-debug:
	docker build --no-cache -t kienghok/scholarship_portal:debug --target debug .

docker-prod:
	docker build --no-cache -t kienghok/scholarship_portal:prod --target prod .


# =========================
# Docker Compose Runs
# =========================
# Use profiles for running debug/prod services
run-debug:
	docker-compose --profile debug up -d --env-file ./.env

run-prod:
	docker-compose --profile prod up -d --env-file ./.env