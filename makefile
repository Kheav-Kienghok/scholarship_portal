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
	@set -a; . ./.env; set +a; goose -dir migrations postgres "$$DATABASE_URL" reset


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