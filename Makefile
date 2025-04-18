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
	@mkdir -p bin

.PHONY: generate
generate: install_yaegi
	@go generate symbols/generate.go

.PHONY: build_amd64
build_amd64: create_bin_dir
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bin/sensor_amd64_linux -trimpath \
	-ldflags "-s -w -X main.BuildCommit=$(GIT_COMMIT) -X main.BuildVersion=$(GIT_TAG) -X main.BuildDate=$(BUILD_DATE) -extldflags=-static" *.go

.PHONY: build_arm64
build_arm64: create_bin_dir
	@CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o bin/sensor_arm64_linux -trimpath \
	-ldflags "-s -w -X main.BuildCommit=$(GIT_COMMIT) -X main.BuildVersion=$(GIT_TAG) -X main.BuildDate=$(BUILD_DATE) -extldflags=-static" *.go

.PHONY: reset_database
reset_database:
	@echo "Resetting database..."
	@rm -r /home/denkhaus/.local/share/sensor

.PHONY: stop_service
stop_service:
	@echo "Stopping service..."
	@sudo systemctl stop sensor.service

.PHONY: start_service
start_service:
	@echo "Starting service..."
	@sudo systemctl start sensor.service

restart_service: stop_service reset_database start_service
	@echo "Restarting service... done"

.PHONY: pull
pull:
	@echo "Pulling latest changes from the repository..."
	@git pull origin master
	@echo "Pulling latest changes from the repository... done"

rebuild_arm: pull build_arm64 restart_service
	@echo "Rebuilding arm64 version and restarting service... done"
