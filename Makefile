APP_NAME = $(shell basename $(shell pwd))

dependencies:
	@echo "Downloading dependencies..."
	@go mod download

lint:
	@echo "Linting..."
	@docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.53.2 golangci-lint run -v

test:
	@echo "Testing..."
	@SKIP_WITH_CREDENTIALS=true go test -v -cover ./...

build:
	@echo "Building..."
	@go build -o bin/$(APP_NAME)


all: dependencies lint test build

.PHONY: dependencies lint test build