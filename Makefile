.PHONY: run build test migrate-up migrate-down seed-admin docker-up docker-prod-up docker-prod-deploy

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server/main.go

seed-admin:
	go run ./cmd/seed/main.go


test:
	go test ./... -v

migrate-up:
	@NETWORK=$$(docker network ls -q -f name=_default | head -1) ; \
	docker run --rm \
	  --network $$NETWORK \
	  -v "$(PWD)/internal/database/migrations:/migrations" \
	  migrate/migrate \
	  -path=/migrations \
	  -database "postgres://bduser:$$(grep -oP 'DB_PASSWORD=\K.*' .env)@postgres:5432/bdgovtjobs?sslmode=disable" up

migrate-down:
	@NETWORK=$$(docker network ls -q -f name=_default | head -1) ; \
	docker run --rm \
	  --network $$NETWORK \
	  -v "$(PWD)/internal/database/migrations:/migrations" \
	  migrate/migrate \
	  -path=/migrations \
	  -database "postgres://bduser:$$(grep -oP 'DB_PASSWORD=\K.*' .env)@postgres:5432/bdgovtjobs?sslmode=disable" down

docker-up:
	docker compose up --build

PROD = docker compose -f docker-compose.yml -f docker-compose.prod.yml

docker-prod-up:
	$(PROD) up -d

docker-prod-deploy:
	$(PROD) pull api
	$(PROD) up -d api

docker-prod-logs:
	$(PROD) logs -f api

docker-prod-down:
	$(PROD) down

scrape:
	curl -X POST http://localhost:8080/api/v1/admin/scrape/run \
	  -H "Authorization: Bearer $(ADMIN_TOKEN)"
