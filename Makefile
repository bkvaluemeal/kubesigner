VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")

UPX := $(shell command -v upx 2> /dev/null)

# Go related variables.
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOFILES := $(shell find cmd -name "*.go")
LDFLAGS=-ldflags "-s -w -X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

install: build
ifdef UPX
	$(UPX) --best --ultra-brute $(GOBIN)/$(PROJECTNAME)
endif

build: go-build

exec:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) $(run)

clean: go-clean
	@-rm $(GOBIN)/$(PROJECTNAME) 2> /dev/null

go-build:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -mod=mod $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

go-clean:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean
