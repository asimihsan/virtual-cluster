APP_NAME := $(shell basename $(PWD))

build-docker:
	docker build -t $(APP_NAME) .