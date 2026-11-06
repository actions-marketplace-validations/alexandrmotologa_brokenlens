.PHONY: all build build-web build-go test clean lint run

BINARY_NAME=brokenlens

all: build

build-web:
	cd web && npm install && npm run build

build-go:
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/brokenlens

build: build-web build-go

test:
	go test -v ./...

clean:
	rm -rf $(BINARY_NAME) $(BINARY_NAME).exe web/dist web/node_modules

run: build
	./$(BINARY_NAME) https://example.com --concurrency 20 --check-external
