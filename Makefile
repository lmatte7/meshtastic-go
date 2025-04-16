.PHONY: default build test lint clean

default: build

build: test
	./build.sh

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf builds
