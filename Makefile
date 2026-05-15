# Makefile — Go via Docker only (host: Docker + GNU make; no local Go toolchain).
SHELL := /bin/bash
.DEFAULT_GOAL := help

GO_VERSION ?= 1.26
DEV_IMAGE ?= edgeos-adblock-dev:latest
DOCKERFILE ?= Dockerfile.dev
E2E_IMAGE ?= edgeos-adblock-e2e:latest
DOCKERFILE_E2E ?= Dockerfile.e2e
MAIN_PKG := ./cmd/update-dnsmasq
EXE := update-dnsmasq

VERBOSE ?= 0
ifeq ($(VERBOSE),1)
  Q :=
else
  Q := @
endif

REPO := $(CURDIR)
GIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
VER ?= $(GIT)
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

DOCKER_BUILD = docker build -f "$(DOCKERFILE)" --build-arg "GO_VERSION=$(GO_VERSION)" -t "$(DEV_IMAGE)" "$(REPO)"
DOCKER_RUN = docker run --rm $(DOCKER_GO_CACHE) -e VER="$(VER)" -v "$(REPO):/src" -w /src "$(DEV_IMAGE)"
DOCKER_BUILD_E2E = docker build -f "$(DOCKERFILE_E2E)" --build-arg "GO_VERSION=$(GO_VERSION)" -t "$(E2E_IMAGE)" "$(REPO)"
DOCKER_RUN_E2E = docker run --rm $(DOCKER_GO_CACHE) -e VER="$(VER)" -v "$(REPO):/src" -w /src "$(E2E_IMAGE)"

.PHONY: help guard-makefile docker-image docker-image-e2e test test-e2e build build-mips64 build-mipsle build-e2e pkgs pkg-mips pkg-mipsel clean

help: ## Show targets (Go runs inside Docker images from this Makefile).
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n"} /^##@/ { printf "\n%s\n", substr($$0, 5) } /^[A-Za-z0-9_.-]+:.*##/ { printf "  %-28s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Setup

docker-image: ## Build the dev image. Optional GO_CACHE_VOLUME=name to isolate the persisted Go mod/build cache volume.
	$(Q) $(DOCKER_BUILD)

docker-image-e2e: ## Build the E2E image (Go + dnsmasq; root parity tests). Optional E2E_IMAGE / DOCKERFILE_E2E.
	$(Q) $(DOCKER_BUILD_E2E)

guard-makefile: ## Assert Makefile conventions (category headers, default goal).
	$(Q) ./scripts/check-makefile-conventions.sh

##@ Build

build: build-mips64 build-mipsle ## Cross-compile linux/mips64 + linux/mipsle → dist/$(EXE).mips and dist/$(EXE).mipsel.
	@echo "Artifacts: $(REPO)/dist/$(EXE).mips $(REPO)/dist/$(EXE).mipsel"

pkgs: pkg-mips pkg-mipsel ## Build EdgeOS .deb for mips + mipsel → dist/ (Vyatta templates + CLI; post-install enables blocklist).
	@echo "Packages in dist/: $$(ls -1 "$(REPO)/dist/"*.deb 2>/dev/null || true)"

pkg-mips: build-mips64 ## Package linux/mips64 binary for EdgeRouter (ER-Lite class) → dist/edgeos-adblock_*_mips.deb
	$(Q) docker image inspect "$(DEV_IMAGE)" >/dev/null 2>&1 || $(DOCKER_BUILD)
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) $(DOCKER_RUN) sh -eu -c '\
		install -m0755 "dist/$(EXE).mips" ".payload/config/scripts/$(EXE)" \
		&& ./make_deb "$(EXE)" mips \
		&& for f in edgeos-adblock_*_mips.deb; do mv -f "$$f" dist/; done \
		&& for f in edgeos-adblock_*_mips.deb.tgz; do test -e "$$f" && mv -f "$$f" dist/ || true; done'

pkg-mipsel: build-mipsle ## Package linux/mipsle binary (ER-X class) → dist/edgeos-adblock_*_mipsel.deb
	$(Q) docker image inspect "$(DEV_IMAGE)" >/dev/null 2>&1 || $(DOCKER_BUILD)
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) $(DOCKER_RUN) sh -eu -c '\
		install -m0755 "dist/$(EXE).mipsel" ".payload/config/scripts/$(EXE)" \
		&& ./make_deb "$(EXE)" mipsel \
		&& for f in edgeos-adblock_*_mipsel.deb; do mv -f "$$f" dist/; done \
		&& for f in edgeos-adblock_*_mipsel.deb.tgz; do test -e "$$f" && mv -f "$$f" dist/ || true; done'

build-mips64: ## linux/mips64 (ER-Lite class) → dist/$(EXE).mips (optional GOMIPS64=softfloat).
	$(Q) docker image inspect "$(DEV_IMAGE)" >/dev/null 2>&1 || $(DOCKER_BUILD)
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) docker run --rm $(DOCKER_GO_CACHE) -v "$(REPO):/src" -w /src \
		-e GOOS=linux -e GOARCH=mips64 -e CGO_ENABLED=0 \
		$(if $(strip $(GOMIPS64)),-e GOMIPS64=$(GOMIPS64),) \
		"$(DEV_IMAGE)" sh -eu -c 'go build -trimpath -mod=readonly \
		-ldflags "$(LDFLAGS) -X main.architecture=mips64 -X main.hostOS=linux $(EXTRA_LDFLAGS)" \
		-o "dist/$(EXE).mips" "$(MAIN_PKG)"'
	@echo "Built $(REPO)/dist/$(EXE).mips (linux/mips64)"
	@file "$(REPO)/dist/$(EXE).mips"
	@ls -l "$(REPO)/dist/$(EXE).mips"

build-mipsle: ## linux/mipsle (ER-X class) → dist/$(EXE).mipsel (default MIPSLE_GOMIPS=softfloat).
	$(Q) docker image inspect "$(DEV_IMAGE)" >/dev/null 2>&1 || $(DOCKER_BUILD)
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) docker run --rm $(DOCKER_GO_CACHE) -v "$(REPO):/src" -w /src \
		-e GOOS=linux -e GOARCH=mipsle -e CGO_ENABLED=0 \
		$(if $(strip $(MIPSLE_GOMIPS)),-e GOMIPS=$(MIPSLE_GOMIPS),) \
		"$(DEV_IMAGE)" sh -eu -c 'go build -trimpath -mod=readonly \
		-ldflags "$(LDFLAGS) -X main.architecture=mipsle -X main.hostOS=linux $(EXTRA_LDFLAGS)" \
		-o "dist/$(EXE).mipsel" "$(MAIN_PKG)"'
	@echo "Built $(REPO)/dist/$(EXE).mipsel (linux/mipsle)"
	@file "$(REPO)/dist/$(EXE).mipsel"
	@ls -l "$(REPO)/dist/$(EXE).mipsel"

build-e2e: ## linux/amd64 → dist/$(EXE).e2e (E2E harness smoke; not for EdgeRouter install).
	$(Q) docker image inspect "$(DEV_IMAGE)" >/dev/null 2>&1 || $(DOCKER_BUILD)
	$(Q) mkdir -p "$(REPO)/dist"
	$(Q) docker run --rm $(DOCKER_GO_CACHE) -v "$(REPO):/src" -w /src \
		-e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 \
		"$(DEV_IMAGE)" sh -eu -c 'go build -trimpath -mod=readonly \
		-ldflags "$(LDFLAGS) -X main.architecture=amd64 -X main.hostOS=linux $(EXTRA_LDFLAGS)" \
		-o "dist/$(EXE).e2e" "$(MAIN_PKG)"'
	@echo "Built $(REPO)/dist/$(EXE).e2e (linux/amd64)"
	@file "$(REPO)/dist/$(EXE).e2e"
	@ls -l "$(REPO)/dist/$(EXE).e2e"

##@ Test

test: ## Run go test ./... in Docker (e.g. TEST_FLAGS=-count=1 make test).
	$(Q) docker image inspect "$(DEV_IMAGE)" >/dev/null 2>&1 || $(DOCKER_BUILD)
	$(Q) $(DOCKER_RUN) sh -eu -c "go mod download && go test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./..."

test-e2e: ## Run go test ./... in E2E image as root (dnsmasq parity tests; needs outbound HTTPS). Uses -count=1 so results are not reused from non-root make test runs.
	$(Q) docker image inspect "$(E2E_IMAGE)" >/dev/null 2>&1 || $(DOCKER_BUILD_E2E)
	$(Q) $(DOCKER_RUN_E2E) sh -eu -c "go mod download && go test -count=1 $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./..."

##@ Clean

clean: ## Remove dist/, cross-build binaries, Debian packages, and common test artefacts.
	$(Q) rm -rf "$(REPO)/dist"
	$(Q) rm -f "$(REPO)"/edgeos-adblock_*.deb "$(REPO)"/edgeos-adblock_*.deb.tgz 2>/dev/null || true
	$(Q) find "$(REPO)" -name "$(EXE).*" -type f -print0 2>/dev/null | xargs -0 rm -f 2>/dev/null || true
	$(Q) find "$(REPO)" -type f \( -name debug -o -name '*.test' -o -name '*.out' \) -print0 2>/dev/null | xargs -0 rm -f 2>/dev/null || true
	$(Q) rm -rf "$(REPO)/test/tests."* "$(REPO)/test/coverage."* 2>/dev/null || true
