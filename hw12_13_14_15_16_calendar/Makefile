BIN_CALENDAR := "./bin/calendar"
BIN_SCHEDULER := "./bin/scheduler"
BIN_SENDER := "./bin/sender"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN_CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

run: build
	$(BIN_CALENDAR) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN_CALENDAR) version

test:
	go test -race ./internal/...

install-lint-deps:
	command -v golangci-lint >/dev/null 2>&1 || \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.64.8 -- -b $(shell go env GOPATH)/bin

lint: install-lint-deps
	PATH=$(shell pwd)/bin:$$PATH golangci-lint run ./...

migrate:
	go run -tags=migrations ./cmd/calendar/migrate/main.go

generate:
	go generate ./grpc/pb/...

.PHONY: build run build-img run-img version test lint generate
