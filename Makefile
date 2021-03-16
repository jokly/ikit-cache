GO ?= go
GOOS ?= $(shell uname -s | tr A-Z a-z)
GOARCH ?= amd64
CGO_ENABLED ?= 0
GOPROXY ?= direct
LDFLAGS = -w -s

GO_ENV = env GOPROXY=$(GOPROXY) GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED)
GO_BUILD ?= $(GO_ENV) $(GO) build -ldflags "$(LDFLAGS)"
GO_RUN ?= $(GO_ENV) $(GO) run
GO_TEST ?= $(GO_ENV) $(GO) test -count=1 -failfast

all: cache consumer

cache:
	$(GO_BUILD) -o cache cmd/cache/*.go

consumer:
	$(GO_BUILD) -o consumer cmd/consumer/*.go

run_cache:
	$(GO_RUN) cmd/cache/*.go

run_consumer:
	$(GO_RUN) cmd/consumer/*.go

clean:
	@rm cache || true
	@rm consumer || true