.PHONY: build install test clean lint deps dev build-all release

BINARY_NAME=passflow
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

install: build
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

test:
	go test -v -race -short ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/ dist/
	rm -f coverage.out coverage.html

lint:
	golangci-lint run ./...

deps:
	go mod download
	go mod tidy

dev:
	go run . $(ARGS)

# Cross-platform builds
build-all: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_windows_amd64.exe .

# Create release archives
release: build-all
	cd dist && \
	tar -czvf $(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz $(BINARY_NAME)_darwin_amd64 && \
	tar -czvf $(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz $(BINARY_NAME)_darwin_arm64 && \
	tar -czvf $(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz $(BINARY_NAME)_linux_amd64 && \
	tar -czvf $(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz $(BINARY_NAME)_linux_arm64 && \
	zip $(BINARY_NAME)_$(VERSION)_windows_amd64.zip $(BINARY_NAME)_windows_amd64.exe && \
	sha256sum $(BINARY_NAME)_$(VERSION)_* > checksums.txt

# Examples:
# make dev ARGS="pal validate agent.yaml"
# make dev ARGS="agents list"
# make release VERSION=v0.1.0
