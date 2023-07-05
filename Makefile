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

.PHONY: example1
example1:
	make build
	rm -f /tmp/vcluster.sqlite3
	build/virtual-cluster substrate start \
		--db-path /tmp/vcluster.sqlite3 \
  		--config-file test_services/http_service_with_kafka/http_service_with_kafka.vcluster \
  		--working-dir 'http_service_with_kafka=./test_services/http_service_with_kafka' \
  		--verbose
