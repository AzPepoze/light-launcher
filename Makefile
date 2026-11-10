.PHONY: all build dev build-go build-renderer clean

all: build

build: build-go build-renderer

build-go:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/light-launcher ./core/cmd/launcher
	go build -ldflags="-s -w" -o bin/light-launcher-instance ./core/cmd/instance

build-renderer:
	@cd src/renderer && bun run build

dev:
	@bun run dev

clean:
	@rm -rf bin dist src/renderer/dist
