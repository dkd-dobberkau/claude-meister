VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build install test clean

build:
	go build -ldflags "-X github.com/dkd-dobberkau/claude-meister/cmd.version=$(VERSION)" -o claude-meister .

install: build
	cp claude-meister $(GOPATH)/bin/ 2>/dev/null || cp claude-meister ~/go/bin/

test:
	go test ./... -v

clean:
	rm -f claude-meister
