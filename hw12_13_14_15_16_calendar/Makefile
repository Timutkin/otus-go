BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	command -v golangci-lint >/dev/null 2>&1 || \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest -- -b $(shell go env GOPATH)/bin

lint: install-lint-deps
	PATH=$(shell pwd)/bin:$$PATH golangci-lint run ./...

migrate:
	go run -tags=migrations ./cmd/calendar/migrate/main.go

.PHONY: build run build-img run-img version test lint
