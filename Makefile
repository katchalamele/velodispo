.PHONY: up down test itest migrate build vet fmt tidy

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
