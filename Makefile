.PHONY: up down test itest migrate build vet fmt tidy swag lint image prod-up prod-down

up:
	docker compose up

down:
	docker compose down

test:
	docker compose run --rm test

itest:
	docker compose up -d db
	docker compose run --rm itest

migrate:
	docker compose run --rm app go run ./cmd/velodispo -migrate

build:
	docker compose run --rm build

vet:
	docker compose run --rm vet

fmt:
	docker compose run --rm fmt

tidy:
	docker compose run --rm dev go mod tidy

swag:
	docker compose run --rm dev go run github.com/swaggo/swag/cmd/swag init -g cmd/velodispo/main.go -o internal/api/docs

lint:
	docker compose run --rm dev go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run

image:
	docker build -t ghcr.io/katchalamele/velodispo:latest .

prod-up:
	docker compose -f compose.prod.yml pull
	docker compose -f compose.prod.yml up -d

prod-down:
	docker compose -f compose.prod.yml down
