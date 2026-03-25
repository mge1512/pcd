.PHONY: build test install clean

build:
	CGO_ENABLED=0 go build -o pcdp-lint .

test:
	go test -v ./...

install: build
	install -m 0755 pcdp-lint /usr/local/bin/

clean:
	rm -f pcdp-lint
	go clean -cache
