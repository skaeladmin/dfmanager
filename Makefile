.DEFAULT_GOAL := build

COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE = `date +%FT%T%z`

GO = go
BINARY_DIR=bin

GODIRS_NOVENDOR = $(shell go list ./... | grep -v /vendor/)
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: test build

help:
	@echo "build      - go build"
	@echo "test       - go test"
	@echo "checkstyle - gofmt+golint+misspell"

get-build-deps:
	$(GO) get github.com/alecthomas/gometalinter
	gometalinter --install

vendor:
	$(if $(shell which glide 2>/dev/null),$(echo "Glide is already installed..."),$(shell go get github.com/Masterminds/glide))
	glide install --strip-vendor

test:
	$(GO) test ${GODIRS_NOVENDOR}

checkstyle:
	gometalinter --vendor ./... --fast --deadline 10m

fmt:
	gofmt -l -w -s ${GOFILES_NOVENDOR}

# Builds chathook
build-service: checkstyle test
	CGO_ENABLED=0 GOOS=linux $(GO) build ${BUILD_INFO_LDFLAGS} -o ${BINARY_DIR}/${service} ./${service}

run:
	realize start --name ${service}

clean:
	if [ -d ${BINARY_DIR} ] ; then rm -r ${BINARY_DIR} ; fi
	if [ -d 'build' ] ; then rm -r 'build' ; fi
