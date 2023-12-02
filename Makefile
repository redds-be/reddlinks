all: compile

check: fmt lint mod vet

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
	@systemctl daemon-reload
	@echo -n "You're almost done, You now need to:\n1. Install postgresql and create a user and a database for rlinks.\n2. Edit /opt/rlinks/.env, either as the root or rlinks user according to the comments\n3. Configure your web server/reverse proxy\n4. Run 'systemctl enable --now rlinks.service'\nAfter that You should be good to go.\n"

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

vet:
	go vet

mod:
	go mod vendor
	go mod tidy

lintall:
	golangci-lint run --enable-all ./...