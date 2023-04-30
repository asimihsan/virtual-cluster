APP_NAME := $(shell basename $(PWD))

build-docker:
	docker buildx build -t $(APP_NAME) .
