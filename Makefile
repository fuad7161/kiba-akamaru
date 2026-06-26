.PHONY: run build test migrate-up migrate-down docker-up

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server/main.go


test:
	go test ./... -v

migrate-up:
	@docker run --rm \
	  --network job-circuler_default \
	  -v "$(PWD)/internal/database/migrations:/migrations" \
	  migrate/migrate \
	  -path=/migrations \
	  -database "postgres://bduser:$$(grep -oP 'DB_PASSWORD=\K.*' .env)@postgres:5432/bdgovtjobs?sslmode=disable" up

migrate-down:
	@docker run --rm \
	  --network job-circuler_default \
	  -v "$(PWD)/internal/database/migrations:/migrations" \
	  migrate/migrate \
	  -path=/migrations \
	  -database "postgres://bduser:$$(grep -oP 'DB_PASSWORD=\K.*' .env)@postgres:5432/bdgovtjobs?sslmode=disable" down

docker-up:
	docker compose up --build

scrape:
	curl -X POST http://localhost:8080/api/v1/admin/scrape/run \
	  -H "Authorization: Bearer $(ADMIN_TOKEN)"
