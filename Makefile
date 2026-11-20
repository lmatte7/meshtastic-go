.PHONY: default build test lint fmt fmt-check proto clean

default: build

build: test
	./build.sh

test:
	go test -v ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || { echo "gofmt needed on:"; gofmt -l .; exit 1; }

# Regenerate the protobuf bindings; needs protoc and protoc-gen-go.
proto:
	./genproto.sh

clean:
	rm -rf builds
