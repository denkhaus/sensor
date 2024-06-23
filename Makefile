GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --tags --abbrev=0)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
########################################################################################
default: build

.PHONY: install_yaegi
install_yaegi:
	go install github.com/traefik/yaegi/cmd/yaegi@latest

.PHONY: create_bin_dir
create_bin_dir:
	mkdir -p bin

.PHONY: generate
generate: install_yaegi
	go generate symbols/generate.go
	
.PHONY: build_amd64
build_amd64: create_bin_dir
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bin/sensor_amd64_linux -trimpath \
	-ldflags "-s -w -X main.BuildCommit=$(GIT_COMMIT) -X main.BuildVersion=$(GIT_TAG) -X main.BuildDate=$(BUILD_DATE) -extldflags=-static" *.go	

.PHONY: build_arm64
build_arm64: create_bin_dir	
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o bin/sensor_arm64_linux -trimpath \
	-ldflags "-s -w -X main.BuildCommit=$(GIT_COMMIT) -X main.BuildVersion=$(GIT_TAG) -X main.BuildDate=$(BUILD_DATE) -extldflags=-static" *.go
