.PHONY: run build test migrate-up migrate-down docker-up

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server/main.go

test:
	go test ./... -v

migrate-up:
	migrate -path internal/database/migrations \
	        -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down:
	migrate -path internal/database/migrations \
	        -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

docker-up:
	docker compose up --build

scrape:
	curl -X POST http://localhost:8080/api/v1/admin/scrape/run \
	  -H "Authorization: Bearer $(ADMIN_TOKEN)"
