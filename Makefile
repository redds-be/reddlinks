GOFILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
VERSION = $(shell git describe --tags)
TESTDB = $(shell find . -type f -name '*_test.db')

all: compile

prep: fmt mod vet lint test

compile: clean
	@mkdir -p build/
	@sed -i "s/noVersion/$(VERSION)/g" main.go
	@go build -o build/

install:
	@useradd rlinks
	@mkdir -p /opt/rlinks
	@chown rlinks:rlinks /opt/rlinks
	@install -o rlinks -g rlinks -m 0755 build/rlinks /opt/rlinks
	@install -o rlinks -g rlinks -m 0600 .env.example /opt/rlinks/.env
	@install -o rlinks -g rlinks -m 0755 rlinks.service /etc/systemd/system/
	@cp -r static/ /opt/rlinks
	@chown -R rlinks:rlinks /opt/rlinks
	@systemctl daemon-reload
	@echo -n "You're almost done, You now need to:\n1. Install postgresql and create a user and a database for rlinks.\n2. Edit /opt/rlinks/.env, either as the root or rlinks user according to the comments\n3. Configure your web server/reverse proxy\n4. Run 'systemctl enable --now rlinks.service'\nAfter that You should be good to go.\n"

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
	go test .

clean-test:
	@for db in $(TESTDB); do rm $$db; done

clean:
	@rm -rf build/