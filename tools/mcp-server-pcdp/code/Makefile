.PHONY: build test install clean container container-podman lint fmt vet coverage

# Build configuration
BINARY_NAME=mcp-server-pcdp
VERSION=0.1.0
LDFLAGS=-ldflags "-X main.serverVersion=$(VERSION)"
CONTAINER_IMAGE=mcp-server-pcdp:latest
CONTAINER_REGISTRY=registry.suse.com

# Build targets
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) .

test:
	go test -v ./...

install: build
	install -D -m 0755 $(BINARY_NAME) $(DESTDIR)/usr/bin/$(BINARY_NAME)
	install -D -m 0644 $(BINARY_NAME).service $(DESTDIR)/usr/lib/systemd/system/$(BINARY_NAME).service

clean:
	rm -f $(BINARY_NAME)
	go clean

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## Container build targets

# Build OCI container using podman
# Uses registry.suse.com/bci/golang:latest as builder base image
# Final stage uses FROM scratch for minimal image size
container: container-podman

container-podman:
	@echo "Building container image with podman..."
	@echo "Builder base: registry.suse.com/bci/golang:latest"
	@echo "Final stage: FROM scratch"
	podman build -t $(CONTAINER_IMAGE) -f Containerfile .
	@echo "Container image built successfully: $(CONTAINER_IMAGE)"
	@echo ""
	@echo "To run the container:"
	@echo "  podman run -p 8080:8080 $(CONTAINER_IMAGE)"
	@echo ""
	@echo "To test the container:"
	@echo "  podman run $(CONTAINER_IMAGE) http listen=127.0.0.1:8080"

container-test: container-podman
	@echo "Testing container image with podman..."
	@echo "Starting container..."
	podman run --rm -d --name mcp-test -p 8080:8080 $(CONTAINER_IMAGE)
	@echo "Container started. Waiting for startup..."
	sleep 2
	@echo "Testing HTTP endpoint..."
	curl -s -X POST http://127.0.0.1:8080/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | jq . || true
	@echo ""
	@echo "Stopping test container..."
	podman stop mcp-test || true
	@echo "Test complete"

container-clean:
	podman rmi -f $(CONTAINER_IMAGE) || true
	@echo "Container image removed"

.DEFAULT_GOAL := build
