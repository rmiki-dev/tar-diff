PROJECT := tar-diff
VERSION := $(shell grep -oP 'VERSION\s*=\s*"\K[^"]+' pkg/common/version.go)
PROJ_TARBALL := $(PROJECT)_$(VERSION).tar.gz

.PHONY: all build clean fmt install lint test tools dist unit-test integration-test validate .install.golangci-lint

export GOPROXY=https://proxy.golang.org


GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
endif

GOFLAGS:=
ifeq ($(GOFLAGS),)
GOFLAGS := -buildvcs=false
endif

PACKAGES := $(shell go list $(GOFLAGS) ./...)
SOURCE_DIRS = $(shell echo $(PACKAGES) | awk 'BEGIN{FS="/"; RS=" "}{print $$4}' | uniq)

PREFIX ?= /usr
INSTALLDIR=${DESTDIR}${PREFIX}/bin

export PATH := $(PATH):${GOBIN}

all: tools tar-diff tar-patch test validate

$(PROJ_TARBALL):
	git archive --prefix=tar-diff_$(VERSION)/ --format=tar.gz HEAD > $(PROJ_TARBALL) 

build:
	go build $(GOFLAGS) ./...

tar-diff:
	go build $(GOFLAGS) ./cmd/tar-diff

tar-patch:
	go build $(GOFLAGS) ./cmd/tar-patch

install: tar-diff tar-patch
	install -d -m 755 ${INSTALLDIR}
	install -m 755 tar-diff ${INSTALLDIR}/tar-diff
	install -m 755 tar-patch ${INSTALLDIR}/tar-patch

tools: .install.golangci-lint

.install.golangci-lint:
	if [ ! -x "$(GOBIN)/golangci-lint" ]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/main/install.sh | sh -s -- -b $(GOBIN) v1.62.2; \
	fi

clean:
	rm -f tar-diff tar-patch
	rm -rf $(PROJECT)_$(VERSION) $(PROJ_TARBALL)

integration-test: tar-diff tar-patch
	tests/test.sh

unit-test:
	go test $(GOFLAGS) -cover ./...

test: unit-test integration-test

fmt:
	@gofmt -l -s -w $(SOURCE_DIRS)

validate: lint
	@go vet $(GOFLAGS) ./...
	@test -z "$$(gofmt -s -l . | tee /dev/stderr)"

lint:
	GOFLAGS=$(GOFLAGS) $(GOBIN)/golangci-lint run

dist: $(PROJ_TARBALL)
	@echo "Created $(PROJ_TARBALL)"

