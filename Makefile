.PHONY: all
all:


build:
	mkdir -p bin
	go build -o bin/terraform-provider-eci

format:
	golangci-lint run ./...
	
	gofmt -w main.go
	gofmt -w internal/api/*.go
	gofmt -w internal/provider/*.go
	gofmt -w internal/resource/*.go

	golines -w main.go
	golines -w internal/api/*.go
	golines -w internal/provider/*.go
	golines -w internal/resource/*.go

check: 
	golangci-lint run ./...

	test -z "$$(gofmt -l main.go)" || (echo "Run 'make format' to fix formatting issues in main.go" && exit 1)
	test -z "$$(gofmt -l internal/api/*.go)" || (echo "Run 'make format' to fix formatting issues in internal/api" && exit 1)
	test -z "$$(gofmt -l internal/provider/*.go)" || (echo "Run 'make format' to fix formatting issues in internal/provider" && exit 1)
	test -z "$$(gofmt -l internal/resource/*.go)" || (echo "Run 'make format' to fix formatting issues in internal/resource" && exit 1)

	test -z "$$(golines -l main.go)" || (echo "Run 'make format' to fix long lines in main.go" && exit 1)
	test -z "$$(golines -l internal/api/*.go)" || (echo "Run 'make format' to fix long lines in internal/api" && exit 1)
	test -z "$$(golines -l internal/provider/*.go)" || (echo "Run 'make format' to fix long lines in internal/provider" && exit 1)
	test -z "$$(golines -l internal/resource/*.go)" || (echo "Run 'make format' to fix long lines in internal/resource" && exit 1)


generate_document:
	tfplugindocs generate --provider-name=eci --examples-dir=examples
	