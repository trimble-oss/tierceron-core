GOPATH=~/workspace/go:$(shell pwd)/vendor:$(shell pwd)
GOBIN=$(shell pwd)/bin
GOFILES=$(wildcard *.go)

ifeq ($(GOOS),)  # Check if GOOS is already set
  GOOS:=$(shell echo $(shell uname -s) | tr '[A-Z]' '[a-z]' | tr -d '[:space:]')
endif

$(info GOOS:$(GOOS))

ifeq ($(GOOS),darwin)
  ifeq ($(shell echo $(shell uname -m) | tr '[A-Z]' '[a-z]'), arm64e)  # Check for 32-bit ARM (armv7l)
    GOARCH := arm64
  else
    GOARCH := amd64
  endif
else ifeq ($(GOOS),linux)
  ifeq ($(shell echo $(shell uname -m) | tr '[A-Z]' '[a-z]'), armv7l)  # Check for 32-bit ARM (armv7l)
    GOARCH := arm
  else ifeq ($(shell echo $(shell uname -m) | tr '[A-Z]' '[a-z]'),aarch64)
    GOARCH := arm64
  else
    GOARCH := amd64  # Assuming 64-bit AMD64 by default for Linux
  endif
else
  $(error Unsupported GOOS: $(GOOS))
endif

$(info GOARCH: $(GOARCH))

fiddler:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) GOOS=$(GOOS) GOARCH=$(GOARCH) go install  -tags "azure memonly"  github.com/trimble-oss/tierceron-core/cmd/trcfiddler

all: fiddler
