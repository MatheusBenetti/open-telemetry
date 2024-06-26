SHELL := /bin/bash

.PHONY: prepare
prepare:
	cp env.test.json env.json

run:
	@docker compose down --rmi local
	@docker compose up -d --force-recreate

.PHONY: init
init:
	go mod tidy

.PHONY: service-a/build
service-a/build:
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o server ./cmd/serviceA

.PHONY: service-b/build
service-b/build:
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o server ./cmd/serviceB