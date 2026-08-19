-include .env

.PHONY: sqlc migrate-up migrate-down migrate-status migrate-test-up migrate-test-down

sqlc:
	sqlc generate

migrate-up:
	goose -dir db/migrations postgres $(DB_URL) up

migrate-down:
	goose -dir db/migrations postgres $(DB_URL) down

migrate-status:
	goose -dir db/migrations postgres $(DB_URL) status

test-migrate-up:
	goose -dir db/migrations postgres $(TEST_DB_URL) up

test-migrate-down:
	goose -dir db/migrations postgres $(TEST_DB_URL) down