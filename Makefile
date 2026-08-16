.PHONY: all build dev build-go build-renderer build-electron clean

all: build

build: build-go build-renderer build-electron

build-go:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/light-launcher-instance ./core/cmd/instance

build-renderer:
	@cd src/renderer && bun run build || npm run build

build-electron:
	@npx tsc -p src/main/tsconfig.json
	@npx tsc -p src/preload/tsconfig.json

dev: build-go
	@npm run dev

dist: build
	@npx electron-builder --config build/electron-builder.yml

clean:
	@rm -rf bin dist src/renderer/dist
