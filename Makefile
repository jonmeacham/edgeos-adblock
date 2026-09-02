# Makefile — Go via the pinned upstream image (host: Docker + GNU make).
SHELL := /bin/bash
.DEFAULT_GOAL := help

GO_IMAGE ?= golang:1.26.3-bookworm@sha256:386d475a660466863d9f8c766fec64d7fdad3edac2c6a05020c09534d71edb4b
MAIN_PKG := ./cmd/edgeos-adblock
EXE := edgeos-adblock

VERBOSE ?= 0
ifeq ($(VERBOSE),1)
  Q :=
else
  Q := @
endif

REPO := $(CURDIR)
GIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
VER ?= 0.0.0+git.$(GIT)
DATE ?= $(shell date +'%FT%H%M%S')
LDFLAGS := -X main.build=$(DATE) -X main.githash=$(GIT) -X main.version=$(VER)
EXTRA_LDFLAGS ?= -s -w

# linux/mipsle (ER-X, etc.): default GOMIPS=softfloat; override with MIPSLE_GOMIPS=hardfloat or MIPSLE_GOMIPS=.
MIPSLE_GOMIPS ?= softfloat

TEST_FLAGS ?=
TEST_TIMEOUT ?= 135s

# Persist Go module and build caches across docker runs.
GO_CACHE_VOLUME ?= edgeos-adblock-go-cache
DOCKER_GO_CACHE = -v "$(GO_CACHE_VOLUME):/cache" -e GOMODCACHE=/cache/mod -e GOCACHE=/cache/go-build
DOCKER_ENV = $(DOCKER_GO_CACHE) -e GOTOOLCHAIN=local -e VER="$(VER)"

DOCKER_RUN = docker run --rm $(DOCKER_ENV) -v "$(REPO):/src" -w /src "$(GO_IMAGE)"
.PHONY: help guard-makefile lint test check build build-mips64 build-mipsle pkgs pkg-mips pkg-mipsel clean

help: ## Show targets (Go runs inside Docker images from this Makefile).
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n"} /^##@/ { printf "\n%s\n", substr($$0, 5) } /^[A-Za-z0-9_.-]+:.*##/ { printf "  %-28s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Setup

guard-makefile: ## Assert Makefile conventions (category headers, default goal).
	$(Q) ./scripts/check-makefile-conventions.sh

##@ Build

build: build-mips64 build-mipsle ## Cross-compile linux/mips64 + linux/mipsle → dist/$(EXE).mips and dist/$(EXE).mipsel.
	@echo "Artifacts: $(REPO)/dist/$(EXE).mips $(REPO)/dist/$(EXE).mipsel"

pkgs: pkg-mips pkg-mipsel ## Build EdgeOS .deb for mips + mipsel → dist/ (Vyatta templates + CLI; post-install enables adblock).
	@echo "Packages in dist/: $$(ls -1 "$(REPO)/dist/"*.deb 2>/dev/null || true)"

pkg-mips: build-mips64 ## Package linux/mips64 binary for EdgeRouter (ER-Lite class) → dist/edgeos-adblock_*_mips.deb
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) $(DOCKER_RUN) sh -eu -c '\
		install -D -m0755 "dist/$(EXE).mips" ".payload/config/scripts/$(EXE)" \
		&& ./make_deb mips \
		&& for f in edgeos-adblock_*_mips.deb; do mv -f "$$f" dist/; done'

pkg-mipsel: build-mipsle ## Package linux/mipsle binary (ER-X class) → dist/edgeos-adblock_*_mipsel.deb
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) $(DOCKER_RUN) sh -eu -c '\
		install -D -m0755 "dist/$(EXE).mipsel" ".payload/config/scripts/$(EXE)" \
		&& ./make_deb mipsel \
		&& for f in edgeos-adblock_*_mipsel.deb; do mv -f "$$f" dist/; done'

build-mips64: ## linux/mips64 (ER-Lite class) → dist/$(EXE).mips (optional GOMIPS64=softfloat).
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) docker run --rm $(DOCKER_ENV) -v "$(REPO):/src" -w /src \
		-e GOOS=linux -e GOARCH=mips64 -e CGO_ENABLED=0 \
		$(if $(strip $(GOMIPS64)),-e GOMIPS64=$(GOMIPS64),) \
		"$(GO_IMAGE)" sh -eu -c 'go build -trimpath -mod=readonly -buildvcs=false \
		-ldflags "$(LDFLAGS) -X main.architecture=mips64 -X main.hostOS=linux $(EXTRA_LDFLAGS)" \
		-o "dist/$(EXE).mips" "$(MAIN_PKG)"'
	@echo "Built $(REPO)/dist/$(EXE).mips (linux/mips64)"
	@file "$(REPO)/dist/$(EXE).mips"
	@ls -l "$(REPO)/dist/$(EXE).mips"

build-mipsle: ## linux/mipsle (ER-X class) → dist/$(EXE).mipsel (default MIPSLE_GOMIPS=softfloat).
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) docker run --rm $(DOCKER_ENV) -v "$(REPO):/src" -w /src \
		-e GOOS=linux -e GOARCH=mipsle -e CGO_ENABLED=0 \
		$(if $(strip $(MIPSLE_GOMIPS)),-e GOMIPS=$(MIPSLE_GOMIPS),) \
		"$(GO_IMAGE)" sh -eu -c 'go build -trimpath -mod=readonly -buildvcs=false \
		-ldflags "$(LDFLAGS) -X main.architecture=mipsle -X main.hostOS=linux $(EXTRA_LDFLAGS)" \
		-o "dist/$(EXE).mipsel" "$(MAIN_PKG)"'
	@echo "Built $(REPO)/dist/$(EXE).mipsel (linux/mipsle)"
	@file "$(REPO)/dist/$(EXE).mipsel"
	@ls -l "$(REPO)/dist/$(EXE).mipsel"

##@ Quality

lint: ## Check Go formatting and run go vet in Docker.
	$(Q) $(DOCKER_RUN) sh -eu -c 'files=$$(gofmt -l cmd internal); \
		if [ -n "$$files" ]; then echo "Go files require formatting:"; echo "$$files"; exit 1; fi; \
		go vet ./...'

##@ Test

test: ## Run go test ./... in Docker (e.g. TEST_FLAGS=-count=1 make test).
	$(Q) $(DOCKER_RUN) sh -eu -c "go test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./..."

check: guard-makefile lint test pkgs ## Run the complete CI verification and package build.

##@ Clean

clean: ## Remove dist/, cross-build binaries, Debian packages, and common test artefacts.
	$(Q) rm -rf "$(REPO)/dist"
	$(Q) rm -f "$(REPO)/.payload/config/scripts/$(EXE)"
	$(Q) rm -f "$(REPO)"/edgeos-adblock_*.deb "$(REPO)"/edgeos-adblock_*.deb.tgz 2>/dev/null || true
	$(Q) find "$(REPO)" -name "$(EXE).*" -type f -print0 2>/dev/null | xargs -0 rm -f 2>/dev/null || true
	$(Q) find "$(REPO)" -type f \( -name debug -o -name '*.test' -o -name '*.out' \) -print0 2>/dev/null | xargs -0 rm -f 2>/dev/null || true
