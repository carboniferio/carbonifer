APP_NAME = $(shell basename $(shell pwd))
VERSION := $(shell git describe --tags)

.PHONY: all

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
	go build -ldflags "-X main.Version=${VERSION}" -o bin/$(APP_NAME)

# Example make tag TAG=v0.0.1
tag:
	@echo "Tagging..."
	@if [ "$(TAG)" == "" ]; then \
        echo "Please set the tag name. For example: make tag TAG=v1.0.0"; \
        exit 1;\
    fi
	git tag -as $(TAG) -m "$(TAG) ($(shell date -u '+%Y-%m-%d %H:%M:%S')))"
	git push upstream $(TAG)

untag:
	@echo "Untagging..."
	@if [ "$(TAG)" == "" ]; then \
        echo "Please set the tag name. For example: make tag TAG=v1.0.0"; \
        exit 1;\
    fi
	git tag -d $(TAG)
	git push upstream :refs/tags/$(TAG)


all: dependencies lint test build

.PHONY: dependencies lint test build