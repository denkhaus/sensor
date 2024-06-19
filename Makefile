GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --tags --abbrev=0)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
########################################################################################
default: build

.PHONY: create_bin_dir
create_bin_dir:
	mkdir -p bin

.PHONY: generate
generate:
	go generate symbols/generate.go
	
.PHONY: build
build: generate create_bin_dir
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bin/sensor_amd64_linux -trimpath \
	-ldflags "-s -w -X main.BuildCommit=$(GIT_COMMIT) -X main.BuildVersion=$(GIT_TAG) -X main.BuildDate=$(BUILD_DATE) -extldflags=-static" *.go

