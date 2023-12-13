GOFILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

all: compile

prep: fmt lint mod vet

compile:
	@mkdir -p build/
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

lint:
	golangci-lint run --enable-all --fix ./...

vet:
	go vet

mod:
	go mod vendor
	go mod tidy

lintall:
	golangci-lint run --enable-all ./...