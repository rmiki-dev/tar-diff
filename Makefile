.PHONY: all build clean fmt install lint test tools unit-test integration-test validate .install.gitvalidation .install.golangci-lint .gitvalidation

export GOPROXY=https://proxy.golang.org

GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(GOPATH)/bin
endif

BUILDFLAGS :=

PACKAGES := $(shell go list $(BUILDFLAGS) ./...)
SOURCE_DIRS = $(shell echo $(PACKAGES) | awk 'BEGIN{FS="/"; RS=" "}{print $$4}' | uniq)

PREFIX ?= ${DESTDIR}/usr
INSTALLDIR=${PREFIX}/bin

export PATH := $(PATH):${GOBIN}

all: tools tar-diff tar-patch test validate .gitvalidation

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

tools: .install.gitvalidation .install.golangci-lint

.install.gitvalidation:
	if [ ! -x "$(GOBIN)/git-validation" ]; then \
		go install github.com/vbatts/git-validation@latest; \
	fi

.install.golangci-lint:
	if [ ! -x "$(GOBIN)/golangci-lint" ]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOBIN) latest; \
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

.gitvalidation:
	@which $(GOBIN)/git-validation > /dev/null 2>/dev/null || (echo "ERROR: git-validation not found. Consider 'make clean && make tools'" && false)
ifeq ($(GITHUB_ACTIONS),true)
	$(GOBIN)/git-validation -q -run DCO,short-subject,dangling-whitespace
else
	git fetch -q "https://github.com/containers/tar-diff.git" "refs/heads/master"
	upstream="$$(git rev-parse --verify FETCH_HEAD)" ; \
		$(GOBIN)/git-validation -q -run DCO,short-subject,dangling-whitespace -range $$upstream..HEAD
endif
