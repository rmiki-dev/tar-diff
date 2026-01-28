.PHONY: all build clean fmt install lint test tools unit-test integration-test validate .install.golangci-lint

export GOPROXY=https://proxy.golang.org


GOBIN ?= $(shell echo $$HOME)/go/bin

export GOFLAGS := -buildvcs=false

BUILDFLAGS :=

PACKAGES := $(shell go list $(BUILDFLAGS) ./...)
SOURCE_DIRS = $(shell echo $(PACKAGES) | awk 'BEGIN{FS="/"; RS=" "}{print $$4}' | uniq)

PREFIX ?= ${DESTDIR}/usr
INSTALLDIR=${PREFIX}/bin

export PATH := $(PATH):${GOBIN}

all: tools tar-diff tar-patch test validate

build:
	go build $(BUILDFLAGS) ./...

tar-diff:
	go build $(BUILDFLAGS) ./cmd/tar-diff

tar-patch:
	go build $(BUILDFLAGS) ./cmd/tar-patch

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

integration-test: tar-diff tar-patch
	tests/test.sh

unit-test:
	go test $(BUILDFLAGS) -cover ./...

test: unit-test integration-test

fmt:
	@gofmt -l -s -w $(SOURCE_DIRS)

validate: lint
	@go vet ./...
	@test -z "$$(gofmt -s -l . | tee /dev/stderr)"

lint:
	$(GOBIN)/golangci-lint run

