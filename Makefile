.PHONY: build build-go run test test-frontend lint clean frontend dev

BINARY := kb
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

frontend:
	cd web && npm ci && npm run build

build: frontend
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/kb

build-go:
	go build -tags dev -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/kb

run: build
	./$(BINARY) serve --open

dev:
	cd web && npm run dev

test:
	go test -tags dev ./... -v -race

test-frontend:
	cd web && npm test

cover:
	go test -tags dev ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY) coverage.out coverage.html
	rm -rf web/dist
