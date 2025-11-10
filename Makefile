.PHONY: all build build-fe build-be

all: build

build: build-fe build-be

build-fe:
	@echo "Building frontend..."
	npm install && npm run build
	@echo "Frontend build complete."

build-be:
	@echo "Building backend..."
	go build -tags netgo -ldflags "-s -w" -o mpd-config-switcher .
	@echo "Backend build complete: ./mpd-config-switcher"
