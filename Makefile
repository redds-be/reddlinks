all: compile

check: fmt lint mod vet

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

vet:
	go vet

mod:
	go mod vendor
	go mod tidy

compile:
	mkdir -p build/
	go build -o build/

lintall:
	golangci-lint run --enable-all ./...