PROJECT_DIR=$(shell pwd)
GOPATH=$(PROJECT_DIR)/libs:$(PROJECT_DIR)

# https://github.com/webrocket/webrocket/blob/master/Makefile

all: build clean

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@echo "  build    to build :)"

clean:
	rm -rf $(PROJECT_DIR)/libs/pkg
	rm -rf $(PROJECT_DIR)/libs/src/*

build:
	GOPATH=$(GOPATH) go build compositor

fmt:
	GOPATH=$(GOPATH) go fmt compositor

doc:
	GOPATH=$(GOPATH) go doc compositor

#get:
#	GOPATH=$(GOPATH) go get github.com/ugorji/go/codec
