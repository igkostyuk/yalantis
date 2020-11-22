SHELL := /bin/bash

export PROJECT = counter

all: counter

counter:
	docker build \
		-f docker/dockerfile.counter-api \
		-t counter-api-amd64:1.0 \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

# ==============================================================================
# Running from within docker compose

run: up 

up:
	docker-compose -f docker/compose/compose.yaml  up --detach --remove-orphans

down:
	docker-compose -f docker/compose/compose.yaml down --remove-orphans

logs:
	docker-compose -f docker/compose/compose.yaml logs -f


test:
	go test ./... -count=1
	staticcheck ./...

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

# ==============================================================================
# Docker support

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

clean:
	docker system prune -f	

logs-local:
	docker logs -f $(FILES)
