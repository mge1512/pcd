.PHONY: build clean test install

BINARY_NAME=pcdp-lint
TEMPLATE_DIR=/usr/share/pcdp/templates/

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -X main.TemplateDir=$(TEMPLATE_DIR)' -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test ./...

install: build
	mkdir -p $(DESTDIR)/usr/bin
	install -m 755 $(BINARY_NAME) $(DESTDIR)/usr/bin/$(BINARY_NAME)
	mkdir -p $(DESTDIR)/usr/share/pcdp/templates

rpm: build
	rpmbuild -ba pcdp-lint.spec

deb: 
	debuild -us -uc