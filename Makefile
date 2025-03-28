default: build

build: test
	go build -o bin/ ./...

test: lint
	go test -v ./...
