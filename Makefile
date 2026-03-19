# ============================================================
#  Onyx — Makefile
# ============================================================

BINARY    := onyx
CMD       := ./cmd/onyx
BUILD_DIR := ./build
DIST_DIR  := ./dist
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-X main.version=$(VERSION) -s -w"
BUILDFLAGS := -buildvcs=false
GO        := go

.PHONY: all build run dev test lint clean build-all fmt vet tidy coverage install

all: lint test build

# ── Build ─────────────────────────────────────────────────────
build:
	@echo "→ Building $(BINARY) $(VERSION)…"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) $(BUILDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "✓ $(BUILD_DIR)/$(BINARY)"

# ── Run / Dev ─────────────────────────────────────────────────
run: build
	$(BUILD_DIR)/$(BINARY) start --dev

dev:
	@which air > /dev/null 2>&1 || (echo "Run: scripts/dev/dev-setup.sh" && exit 1)
	air

# ── Test ──────────────────────────────────────────────────────
test:
	@echo "→ Running tests…"
	$(GO) test ./... -v -race -cover -coverprofile=coverage.out
	@echo "✓ Tests passed"

test-short:
	$(GO) test ./... -short

coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ coverage.html"

# ── Quality ───────────────────────────────────────────────────
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Run: scripts/dev/dev-setup.sh" && exit 1)
	golangci-lint run ./...
	@echo "✓ Lint passed"

fmt:
	$(GO) fmt ./...
	goimports -w . 2>/dev/null || true

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

# ── Install ───────────────────────────────────────────────────
install: build
	sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "✓ Installed to /usr/local/bin/$(BINARY)"

# ── Cross-compile ─────────────────────────────────────────────
build-all:
	@echo "→ Building for all platforms…"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux   GOARCH=amd64  $(GO) build $(LDFLAGS) $(BUILDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64      $(CMD)
	GOOS=linux   GOARCH=arm64  $(GO) build $(LDFLAGS) $(BUILDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64      $(CMD)
	GOOS=darwin  GOARCH=amd64  $(GO) build $(LDFLAGS) $(BUILDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64     $(CMD)
	GOOS=darwin  GOARCH=arm64  $(GO) build $(LDFLAGS) $(BUILDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64     $(CMD)
	GOOS=windows GOARCH=amd64  $(GO) build $(LDFLAGS) $(BUILDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe $(CMD)
	@echo "✓ All binaries in $(BUILD_DIR)/"

# ── Clean ─────────────────────────────────────────────────────
clean:
	@rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html
	@echo "✓ Clean"
