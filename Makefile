.PHONY: build run test test-integration test-e2e lint fmt clean docker

BINARY=ssht
BUILD_DIR=bin
IMAGE_NAME=ssht

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/ssht

run:
	go run ./cmd/ssht

run-debug:
	go run ./cmd/ssht -debug

test:
	go test ./...

test-integration:
	go test -tags=integration -v ./internal/tunnel/...

test-e2e:
	go test -tags=e2e -v ./e2e-test/...

test-e2e-cleanup:
	cd e2e-test && docker compose down -v

test-all: test test-integration

lint:
	golangci-lint run

fmt:
	go fmt ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY) $(GOPATH)/bin/

docker:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run -p 6001:6001 -v ~/.ssh:/root/.ssh:ro -v ~/.config/ssht:/root/.config/ssht $(IMAGE_NAME)
