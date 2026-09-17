.PHONY: build run test test-integration test-e2e lint fmt clean docker web-install web-dev web-build build-all

BINARY=ssht
BUILD_DIR=bin
IMAGE_NAME=ssht

# Go build
build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/ssht

run:
	go run ./cmd/ssht

run-debug:
	go run ./cmd/ssht -debug

# Tests
test:
	go test ./...

test-integration:
	go test -tags=integration -v ./internal/tunnel/...

test-e2e:
	go test -tags=e2e -v ./e2e-test/...

test-e2e-cleanup:
	cd e2e-test && docker compose down -v

test-all: test test-integration

# Code quality
lint:
	golangci-lint run

fmt:
	go fmt ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR) cmd/ssht/dist

# Web frontend
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# Full build (frontend + backend)
build-all: web-build build

# Install
install: build
	cp $(BUILD_DIR)/$(BINARY) $(GOPATH)/bin/

# Docker
docker: web-build
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run -p 6001:6001 \
		-v ~/.ssh:/root/.ssh:ro \
		-v ~/.config/ssht:/root/.config/ssht:ro \
		-v ~/.local/state/ssht:/root/.local/state/ssht \
		$(IMAGE_NAME)

docker-deploy: docker
	docker run -d --name ssht --restart unless-stopped \
		-p 6001:6001 \
		-v ~/.ssh:/root/.ssh:ro \
		-v ~/.config/ssht:/root/.config/ssht:ro \
		-v ~/.local/state/ssht:/root/.local/state/ssht \
		$(IMAGE_NAME)
