PROJECT_NAME := auth-service

.PHONY: all build test clean docker-up docker-down

all: build

build:
	go build -o $(PROJECT_NAME) cmd/api/main.go

test:
	go test ./... -v

clean:
	rm -f $(PROJECT_NAME)

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

run:
	go run cmd/api/main.go
