.PHONY: build test install clean

# Build target - static binary as required by cli-tool template
build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pcdp-lint .

# Test target
test:
	go test -v ./...

# Install target
install: build
	install -m 755 pcdp-lint $(DESTDIR)/usr/bin/

# Clean target
clean:
	rm -f pcdp-lint

# Additional targets for development
fmt:
	go fmt ./...

vet:
	go vet ./...

# Cross-compilation targets
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o pcdp-lint.exe .

build-macos:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o pcdp-lint-darwin .

build-all: build build-windows build-macos