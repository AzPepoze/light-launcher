.PHONY: all build build-go build-shell dev test test-go test-ui e2e lint clean

VERSION ?= $(patsubst v%,%,$(shell git describe --tags --always 2>/dev/null || echo dev))
LDFLAGS := -s -w -X main.version=$(VERSION)

all: build

# build makes the launcher, the instance binary and the Qt shell in bin/.
build: build-go build-shell

build-go:
	@mkdir -p bin
	bash scripts/build-launcher.sh "$(CURDIR)/bin/light-launcher"
	go build -ldflags="$(LDFLAGS)" -o bin/light-launcher-instance ./core/cmd/instance

build-shell:
	@mkdir -p bin
	cd shell/qt && go build -ldflags="$(LDFLAGS)" -o ../../bin/light-launcher-shell .

# dev starts the Vite dev server and the launcher API on port 9246.
dev:
	bun run dev

# test runs the UI tests (vitest) and the Go tests.
test: test-ui test-go

test-ui:
	bun run test

test-go:
	go test -count=1 ./core/...

# e2e starts the built launcher without a window and calls its API. It needs no display.
e2e:
	go test -tags e2e -count=1 ./core/cmd/launcher/...

lint:
	bun run lint
	go vet ./core/...

clean:
	@rm -rf bin dist src/renderer/dist core/cmd/launcher/ui
