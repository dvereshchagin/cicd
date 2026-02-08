APP_NAME := microservice
BINARY := bin/$(APP_NAME)

.PHONY: fmt format lint test build run ci lambda-build lambda-package deploy-lambda

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

lambda-build:
	mkdir -p .build/lambda
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o .build/lambda/bootstrap ./cmd/lambda

lambda-package: lambda-build
	cd .build/lambda && zip -q -j function.zip bootstrap

deploy-lambda:
	./scripts/deploy_lambda.sh

run:
	go run ./cmd/microservice

ci: fmt lint test build
