APP_NAME := $(shell basename $(PWD))

.PHONY: build
build:
	scripts/generate.sh
	go build -o build/virtual-cluster cmd/virtual-cluster/main.go

.PHONY: generate
generate:
	scripts/generate.sh

.PHONY: test
test:
	go test -p 1 ./...
