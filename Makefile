GOFILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
VERSION = $(shell git describe --tags)
TESTDB = $(shell find . -type f -name '*_test.db')

all: compile

prep: fmt mod vet lint test

compile: clean
	@mkdir -p build/
	@go build -ldflags="-X 'main.Version=$(VERSION)'" -o build/

fmt:
	golines --max-len=120 --base-formatter=gofumpt -w $(GOFILES)

mod:
	go mod vendor
	go mod tidy

vet:
	go vet

lint:
	golangci-lint run --enable-all --fix ./...

test: clean-test
	go test ./...

clean-test:
	@for db in $(TESTDB); do rm $$db; done

clean:
	@rm -rf build/
