include .env
export

.PHONY: migrate-up migrate-down migrate-status sqlc-generate

migrate-up:
	cd server && go tool goose -dir db/migrations postgres "$(DB_URL_GOOSE)" up

migrate-down:
	cd server && go tool goose -dir db/migrations postgres "$(DB_URL_GOOSE)" down

migrate-status:
	cd server && go tool goose -dir db/migrations postgres "$(DB_URL_GOOSE)" status

sqlc-generate:
	cd server && go generate ./...