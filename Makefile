APP_NAME := microservice
BINARY := bin/$(APP_NAME)

.PHONY: fmt format lint test build run ci

fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

format:
	@gofmt -w $$(find . -type f -name '*.go')

lint:
	go vet ./...

test:
	go test ./...

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/microservice

run:
	go run ./cmd/microservice

ci: fmt lint test build
