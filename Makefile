# Load environment variables from .env file if it exists
-include .env

.PHONY: sqlc migrate-up migrate-down migrate-status

sqlc:
	sqlc generate

migrate-up:
	goose -dir db/migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir db/migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir db/migrations postgres "$(DB_URL)" status