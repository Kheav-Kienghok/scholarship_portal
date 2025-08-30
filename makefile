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
