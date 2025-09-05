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


# Build docker images
docker-debug:
	docker build -t kienghok/scholarship_portal:debug --target debug .

docker-prod:
	docker build -t kienghok/scholarship_portal:prod --target prod .

# Build and run in one command
run-debug:
	@if ! docker image inspect kienghok/scholarship_portal:debug > /dev/null 2>&1; then \
		$(MAKE) docker-debug; \
	fi
	docker-compose up debug

run-prod:
	@if ! docker image inspect kienghok/scholarship_portal:prod > /dev/null 2>&1; then \
		$(MAKE) docker-prod; \
	fi
	docker-compose up prod
