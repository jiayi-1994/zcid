.PHONY: dev build test swag lint clean migrate-up migrate-down migrate-new fmt vet frontend-install frontend-build frontend-test frontend-lint coverage docker-build all

dev:
	go run cmd/server/main.go

build:
	go build -o bin/zcid cmd/server/main.go

test:
	go test ./... -v

swag:
	swag init -g cmd/server/main.go -o docs

lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

clean:
	rm -rf bin/

migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-new:
	@if [ -z "$(name)" ]; then echo "usage: make migrate-new name=create_xxx"; exit 1; fi
	go run cmd/migrate/main.go new --name "$(name)"

fmt:
	go fmt ./...

vet:
	go vet ./...

frontend-install:
	cd web && npm install

frontend-build:
	cd web && npm run build

frontend-test:
	cd web && npx vitest run

frontend-lint:
	cd web && npx tsc --noEmit

coverage:
	go test ./pkg/... ./internal/... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

docker-build:
	docker build -t zcid:latest .

all: build frontend-build
