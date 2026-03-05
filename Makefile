.PHONY: dev build test swag lint clean migrate-up migrate-down migrate-new

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
