.PHONY: build run test lint docker migrate-up migrate-down

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

docker:
	docker build -t gondor-event-bus .

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down
