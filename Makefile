.PHONY: all build dev build-go build-renderer build-electron dist clean

all: build

build: build-go build-renderer build-electron

build-go:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/light-launcher-instance ./core/cmd/instance

build-renderer:
	@cd src/renderer && bun run build

build-electron:
	@bunx tsc -p src/main/tsconfig.json
	@bunx tsc -p src/preload/tsconfig.json

dev: build-go
	@bun run dev

dist: build
	@bunx electron-builder --config build/electron-builder.yml

clean:
	@rm -rf bin dist src/renderer/dist
