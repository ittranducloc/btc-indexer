VERSION ?= $(shell git rev-parse --short HEAD)

.PHONY: build-local
build:
	@echo "Building app ..."
	@/bin/bash ./bin/build.sh -c -v

.PHONY: generate-builder
builder:
	@echo "Making builder ..."
	@docker build -t builder ./docker/builder/

.PHONY: build-alpine
build-alpine: builder
	@echo "Building app for alpine OS ..."
	@docker run --rm -v $(shell pwd):/go/src/github.com/darkknightbk52/btc-indexer/ -w /go/src/github.com/darkknightbk52/btc-indexer/ builder make build

.PHONY: build-docker
build-docker: build-alpine
	@echo "Building docker image ..."
	@for d in ./docker/*; do \
		TAG="$$(echo "$$(echo $$d | awk -F '/' '{print $$NF}')":${VERSION})"; \
		docker build -t $$TAG $$d; \
	done