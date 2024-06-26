EXECUTABLE ?= grafana-annotations-bot
GO := CGO_ENABLED=0 go
APP_VERSION ?= dev
VPREFIX := github.com/zt-sv/grafana-annotations-bot/internal/pkg/build
LDFLAGS += -X $(VPREFIX).Version=$(APP_VERSION) -X $(VPREFIX).BuildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS += -extldflags '-static'

PACKAGES = $(shell go list ./... | grep -v /vendor/)

.PHONY: all
all: build

.PHONY: clean
clean:
	$(GO) clean -i ./...
	rm -rf dist/

.PHONY: fmt
fmt:
	$(GO) fmt $(PACKAGES)

.PHONY: vet
vet:
	$(GO) vet $(PACKAGES)

.PHONY: lint
lint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) install golang.org/x/lint/golint@latest; \
	fi
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

.PHONY: dep
dep:
	$(GO) mod vendor

.PHONY: build
build:
	$(GO) build -v -ldflags '-w $(LDFLAGS)' -o dist/$(EXECUTABLE) ./cmd/$(EXECUTABLE)

.PHONY: release
release:
	@which gox > /dev/null; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/mitchellh/gox@v1.0.1; \
	fi
	CGO_ENABLED=0 gox -arch="amd64 arm" -verbose -ldflags '-w $(LDFLAGS)' -output="dist/$(EXECUTABLE)-${APP_VERSION}-{{.OS}}-{{.Arch}}" ./cmd/$(EXECUTABLE)
