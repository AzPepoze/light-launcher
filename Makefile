.PHONY: all build dev build-go build-renderer build-electron dist clean

all: build

build: build-go build-renderer build-electron

build-go:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/light-launcher-instance ./core/cmd/instance

build-renderer:
	@cd src/renderer && bun run build

build-electron:
	@bun run build:electron

dev: build-go
	@bun run dev

dist: build
	@bun run dist

clean:
	@rm -rf bin dist src/renderer/dist
