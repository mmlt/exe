.PHONY: all install-tools generate test bin check

VERSION?=v0.0.1

# CI/CD target.
all: generate bin

# Run unit tests.
test: check
	go test ./... -coverprofile test.cover

# Create binaries.
bin: check
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o bin/exerecorder-$(VERSION)-linux-amd64 github.com/mmlt/exe/cmd

# Check code for issues.
check:
	go fmt ./...
	go vet ./...

# Install binary in PATH.
install-linux:
	sudo cp bin/exerecorder-$(VERSION)-linux-amd64 /usr/local/bin/
	sudo ln -sfr /usr/local/bin/exerecorder-$(VERSION)-linux-amd64 /usr/local/bin/exerecorder

