# .env dagi o'zgaruvchilar barcha targetlarga export qilinadi.
include .env
export

.PHONY: up up/full down logs run/auth run/course run/enrollment build vet

## up: dev rejim — db + gateway docker'da, servislar hostda (make run/...)
up:
	docker compose --profile dev up -d

## up/full: butun tizim docker'da (db, 3 servis, gateway, frontend)
up/full:
	docker compose --profile full up -d --build

down:
	docker compose --profile dev --profile full down

logs:
	docker compose --profile dev --profile full logs -f

## Dev rejimda servislar hostda ishlaydi (har biri alohida terminalda)
run/auth:
	cd auth-service && go run ./cmd/api

run/course:
	cd course-service && go run ./cmd/api

run/enrollment:
	cd enrollment-service && go run ./cmd/api

build:
	cd pkg && go build ./...
	cd auth-service && go build ./...
	cd course-service && go build ./...
	cd enrollment-service && go build ./...

vet:
	cd pkg && go vet ./...
	cd auth-service && go vet ./...
	cd course-service && go vet ./...
	cd enrollment-service && go vet ./...
