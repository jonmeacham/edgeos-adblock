# Makefile to build dnsmasq blocklist (canonical interface: make help)
PACKAGE 	= github.com/jonmeacham/edgeos-adblock
SHELL		= /bin/bash

# Go parameters
BASE			 = $(CURDIR)/
BIN 			 = /usr/local/go/bin
GO				 = go
GOPATH_BIN		 = $(shell $(GO) env GOPATH)/bin
GOLINT			 = $(GOPATH_BIN)/golangci-lint
DISTDIR			 = /tmp
GOBUILD			 = $(GO) build -mod=readonly
GOCLEAN			 = $(GO) clean -cache
GODOC			 = godoc
GOFMT			 = gofmt
GOGEN			 = $(GO) generate
GOGET			 = $(GO) install
GOSHADOW		 = $(GO) tool vet -shadow
GOTEST			 = $(GO) test
PKGS	 		 = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./...))
SRC				 = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
TESTPKGS		 = $(shell env GOPATH=$(GOPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
TIMEOUT			 = 135
# Extra flags for go test (e.g. TEST_FLAGS=-count=1 for uncached runs).
TEST_FLAGS		 ?=

# Executable and package variables
EXE				 = update-dnsmasq
TARGET			 = edgeos-adblock
MAIN_PKG		 = ./cmd/update-dnsmasq

# Executables
GSED			 = $(shell which gsed || which sed) -i.bak -e

# Version for ldflags, deb names, and release tags (override: make build VER=1.2.3)
VER			?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0-dev)
OLDVER			?= $(VER)

# Environment variables
AWS				 = aws
COPYRIGHT		 = s/Copyright © 20../Copyright © $(shell date +"%Y")/g
DATE			 = $(shell date +'%FT%H%M%S')
FLAGS 			 = -s -w
GIT				 = $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS 		 = -X main.build=$(DATE) -X main.githash=$(GIT) -X main.version=$(VER)
LIC			 	 = license
PAYLOAD 		 = ./.payload
README 			 = README.md
SCRIPTS 		 = /config/scripts
TAG 			 = v$(VER)

ifeq ("$(origin V)", "command line")
  KBUILD_VERBOSE = $(V)
endif
ifndef KBUILD_VERBOSE
  KBUILD_VERBOSE = 0
endif

ifeq ($(KBUILD_VERBOSE),1)
  quiet =
  Q =
else
  quiet=quiet_
  Q = @
endif
export quiet Q KBUILD_VERBOSE

.DEFAULT_GOAL := help

.PHONY: \
	help \
	install deps \
	build build-local build-ci build-mipsle build-cross-mipsle docker-build \
	generate amd64 arm64 linux mips mips64 mipsle mac \
	all AllOfIt \
	docs readme copyright pkgs pkg-mips pkg-mipsel \
	format-check fmt simplify vet lint shadow report \
	test test-bench test-short test-verbose test-race check tests test-xml profile \
	coverage test-coverage test-coverage-tools \
	guard-makefile ci verify \
	run-app refresh-data deploy-app \
	version release commit push repo upload tags \
	clean

help: ## Show targets grouped by category (see docs/makefile-conventions.md).
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n"} /^##@/ { printf "\n%s\n", substr($$0, 5) } /^[A-Za-z0-9_.-]+:.*##/ { printf "  %-38s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Setup

install: ## Download modules, verify, tidy; install golangci-lint if missing (no vendor/ — use Docker for reproducible builds).
	$(Q) $(GO) mod download
	$(Q) $(GO) mod verify
	$(Q) $(GO) mod tidy
	$(Q) test -x "$(GOLINT)" || $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

deps: install ## Legacy alias for install.

$(GOLINT): ## Install golangci-lint into GOPATH bin (used by lint / ci).
	$(Q) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

GOTOOLS			= $(BIN)/stringer
$(BIN)/stringer: ; @ $(info $(M)  installing stringer…) @ ##   [optional] Install stringer to GOROOT bin
	$(Q) $(GOGET) golang.org/x/tools/...@latest

GODOC2MD		= $(BIN)/godocdown
$(BIN)/godocdown: ; @ $(info $(M)  installing godocdown…) @ ##   [optional] Install godocdown
	$(Q) $(GOGET) github.com/robertkrimen/godocdown/godocdown@latest

GOLINT_LEGACY	 = $(BIN)/golangci-lint
$(BIN)/golangci-lint: ; @ $(info $(M)  installing golangci-lint…) @ ##   [optional] Install golangci-lint to GOROOT bin
	$(Q) $(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

GOCOVMERGE 		 = $(BIN)/gocovmerge
$(BIN)/gocovmerge: ; @ $(info $(M)  installing gocovmerge…) @ ##   [optional] Install gocovmerge
	$(Q) $(GOGET) github.com/wadey/gocovmerge@latest

GOCOV 			 = $(BIN)/gocov
$(BIN)/gocov: ; @ $(info $(M)  installing gocov…) @ ##   [optional] Install gocov
	$(Q) $(GOGET) github.com/axw/gocov/...@latest

GOCOVXML 		 = $(BIN)/gocov-xml
$(BIN)/gocov-xml: ; @ $(info $(M)  installing gocov-xml…) @ ##   [optional] Install gocov-xml
	$(Q) $(GOGET) github.com/AlekSi/gocov-xml@latest

GO2XUNIT 		 = $(BIN)/go2xunit
$(BIN)/go2xunit: ; @ $(info $(M)  installing go2xunit…) @ ##   [optional] Install go2xunit
	$(Q) $(GOGET) github.com/tebeka/go2xunit@latest

GOREPORTER		 = $(BIN)/goreporter
$(BIN)/goreporter: ; @ $(info $(M)  installing goreporter…) @ ##   [optional] Install goreporter
	$(Q) $(GOGET) github.com/360EntSecGroup-Skylar/goreporter@latest

##@ Build

build-local: install ## Build $(EXE) for the current OS/arch (uses committed stringer output; run generate when enums change).
	$(Q) $(GOBUILD) -o $(EXE) -ldflags "$(LDFLAGS) $(FLAGS)" $(MAIN_PKG)

build-ci: install ## Fast compile-only check (writes $(DISTDIR)/$(TARGET).ci).
	$(Q) mkdir -p $(DISTDIR)
	$(Q) $(GOBUILD) -o $(DISTDIR)/$(TARGET).ci $(MAIN_PKG)

build-mipsle: mipsle ## Cross-compile linux/mipsle (EdgeOS ER-X); output $(EXE).mipsel (runs go generate).

# Cross-compile mipsle without "go generate" (avoids host toolchain vs golang.org/x/tools mismatch during stringer).
build-cross-mipsle: install ## Cross-compile linux/mipsle to $(DISTDIR) (no go generate; uses module cache).
	$(Q) mkdir -p $(DISTDIR)
	$(Q) GOOS=linux GOARCH=mipsle CGO_ENABLED=0 $(GOBUILD) \
		-ldflags "$(LDFLAGS) $(FLAGS) -X main.architecture=mipsle -X main.hostOS=linux" \
		-o $(DISTDIR)/$(EXE).mipsle.ci $(MAIN_PKG)

docker-build: ## Build OCI image (compile-only in Docker; run `make ci` on host for tests). Tag: edgeos-adblock:local
	$(Q) docker build -t edgeos-adblock:local .

generate: ; @ $(info $(M) generating go boilerplate code…) ## Run go generate for all packages.
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./...); do \
		cd $$d ; $(GOGEN) || ret=$$? ; \
	done ; exit $$ret

amd64: generate ; @ $(info building darwin/amd64…) ## Cross-compile darwin/amd64 → $(EXE).amd64
	$(eval LDFLAGS += -X main.architecture=amd64 -X main.hostOS=darwin)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(EXE).amd64 \
		-ldflags "$(LDFLAGS) $(FLAGS)" -v $(MAIN_PKG)

arm64: generate ; @ $(info building darwin/arm64…) ## Cross-compile darwin/arm64 → $(EXE).arm64
	$(eval LDFLAGS += -X main.architecture=arm64 -X main.hostOS=darwin)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(EXE).arm64 \
		-ldflags "$(LDFLAGS) $(FLAGS)" -v $(MAIN_PKG)

linux: generate ; @ $(info building linux/amd64…) ## Cross-compile linux/amd64 → $(EXE).linux
	$(eval LDFLAGS += -X main.architecture=amd64 -X main.hostOS=linux)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(EXE).linux \
		-ldflags "$(LDFLAGS) $(FLAGS)" -v $(MAIN_PKG)

mips: mips64 mipsle ; @ $(info building MIPS/MIPSLE binaries…) ## Build mips64 + mipsle router binaries.

mips64: generate ; @ $(info building linux/mips64…) ## Cross-compile linux/mips64 → $(EXE).mips
	$(eval LDFLAGS += -X main.architecture=mips64 -X main.hostOS=linux)
	GOOS=linux GOARCH=mips64 $(GOBUILD) -o $(EXE).mips \
		-ldflags "$(LDFLAGS) $(FLAGS)" -v $(MAIN_PKG)

mipsle: generate ; @ $(info building linux/mipsle…) ## Cross-compile linux/mipsle → $(EXE).mipsel
	$(eval LDFLAGS += -X main.architecture=mipsle -X main.hostOS=linux)
	GOOS=linux GOARCH=mipsle CGO_ENABLED=0 $(GOBUILD) -o $(EXE).mipsel \
		-ldflags "$(LDFLAGS) $(FLAGS)" -v $(MAIN_PKG)

mac: amd64 arm64 ## Build darwin amd64 + arm64 binaries.

build: clean amd64 linux mips copyright docs readme ; @ $(info full release cross-build…) ## Full cross-build for release (darwin amd64, linux, mips variants + docs).
docs: readme $(GODOC2MD) ; @ $(info $(M) building docs…) ## Prepare docs assets (readme hook). Optional: godocdown via $(GODOC2MD).
	@true

readme: ## README.md is maintained in-repo (no template file).
	@true

copyright: ; @ $(info updating copyright…) ## Stamp copyright year in README and license copies.
	$(GSED) '$(COPYRIGHT)' $(README)
	$(GSED) '$(COPYRIGHT)' $(LIC)
	cp $(LIC) internal/edgeos/
	cp $(LIC) internal/regx/
	cp $(LIC) internal/tdata/

pkgs: pkg-mips pkg-mipsel ; @ $(info building Debian packages…) ## Build Debian .deb packages (mips + mipsel).

pkg-mips: deps mips coverage copyright docs readme ; @ $(info building MIPS Debian package…) ## Package deb for mips64 routers.
	cp $(EXE).mips $(PAYLOAD)$(SCRIPTS)/$(EXE) \
	&& ./make_deb $(EXE) mips

pkg-mipsel: deps mipsle coverage copyright docs readme ; @ $(info building MIPSLE Debian package…) ## Package deb for mipsle routers.
	cp $(EXE).mipsel $(PAYLOAD)$(SCRIPTS)/$(EXE) \
	&& ./make_deb $(EXE) mipsel

all: AllOfIt ; @ $(info making everything…) ## Clean, build cross-arch, coverage, docs, deb packages.
AllOfIt: clean deps amd64 mips coverage copyright docs readme pkgs

##@ Quality

format-check: ## Fail if gofmt would change any tracked .go file (skips ./vendor if present).
	@files="$$(gofmt -l $$(find . -type f -name '*.go' -not -path './vendor/*'))"; \
	if [ -n "$$files" ]; then echo >&2 "$$files"; exit 1; fi

fmt: ; $(info $(M) running gofmt…) @ ## Rewrite all non-vendor .go files with gofmt.
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./...); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	done ; exit $$ret

simplify: ; @ $(info simplifying code…) ## Apply gofmt -s (simplify) in place.
	@$(GOFMT) -s -l -w $(SRC)

vet: install ## Run go vet on all packages.
	$(Q) cd $(BASE) && $(GO) vet ./...

lint: install $(GOLINT) | $(BASE) ; $(info $(M) running golangci-lint…) @ ## Run golangci-lint using .golangci.yml.
	$(Q) "$(GOLINT)" run ./...

shadow: ; $(info $(M) running go vet -shadow…) @ ## Check for variable shadowing (legacy vet mode).
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./...); do \
		$(GOSHADOW) $$d/*.go || ret=$$? ; \
	done ; exit $$ret

report: ; $(info $(M) running goreporter…) @ ## HTML quality report (requires $(GOREPORTER)).
	$(GOREPORTER) -p $(CURDIR) -f html \
	-e "vendor/golang.org"

##@ Test

test-bench: ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks only.
test-short: ARGS=-short ## Run only tests marked short.
test-verbose: ARGS=-v ## Verbose test output.
test-race: ARGS=-race ## Tests with the race detector (amd64 recommended).

TEST_TARGETS := test-bench test-short test-verbose test-race

test $(TEST_TARGETS): ## Run go test ./... (variants: test-short, test-race; matches CI compile scope).
	$(Q) cd $(BASE) && $(GO) test $(TEST_FLAGS) -timeout $(TIMEOUT)s $(ARGS) ./...

check tests: fmt install | $(BASE) ; $(info $(M) running fmt + install + tests…) @ ## fmt + modules + test ./... (fmt rewrites files).
	$(Q) cd $(BASE) && $(GO) test $(TEST_FLAGS) -timeout $(TIMEOUT)s $(ARGS) ./...

test-xml: fmt lint install | $(BASE) $(GO2XUNIT) ; $(info $(M) tests → JUnit XML…) @ ## Tests → test/tests.xml (needs go2xunit).
	$(Q) cd $(BASE) && 2>&1 $(GO) test $(TEST_FLAGS) -timeout $(TIMEOUT)s -v ./... | tee test/tests.output
	$(Q) $(GO2XUNIT) -fail -input test/tests.output -output test/tests.xml

profile: ; $(info $(M) profiling code…) @ ## CPU/mem profiles per package (development).
	$(Q) cd $(BASE)
	$(foreach pkg,$(TESTPKGS), $(shell [[ $(notdir ${pkg}) != "tdata" ]] && $(GO) test -cpuprofile cpu.$(notdir ${pkg}).prof -memprofile mem.$(notdir ${pkg}).prof) -bench ${pkg})

##@ Coverage

COVERAGE_MODE    = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML     = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML    = $(COVERAGE_DIR)/index.html

coverage: test-coverage ; $(info $(M) coverage alias…) @ ## Alias for test-coverage.

test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML)

test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: fmt lint install test-coverage-tools | $(BASE) ; $(info $(M) generating merged coverage…) @ ## Per-package coverage, HTML + Cobertura XML under test/coverage.<timestamp>/.
	$(Q) mkdir -p $(COVERAGE_DIR)/coverage
	$(Q) cd $(BASE) && for pkg in $(TESTPKGS); do \
		$(GO) test \
			-coverpkg=$$($(GO) list -f '{{ join .Deps "\n" }}' $$pkg | \
					grep '^$(PACKAGE)/' | grep -v '^$(PACKAGE)/vendor/' | \
					tr '\n' ',')$$pkg \
			-covermode=$(COVERAGE_MODE) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$(Q) $(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	$(Q) $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$(Q) $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

##@ CI

guard-makefile: ## Assert Makefile conventions (default goal is help; Setup→Clean category headers).
	$(Q) ./scripts/check-makefile-conventions.sh

# Order: guard → deps → static quality → tests (-count=1) → compile checks (host + mipsle).
ci: guard-makefile install format-check vet lint ## Full gate: deps, quality, uncached tests, host + mipsle compile (see docs/build-and-test.md).
	$(Q) $(MAKE) -f $(firstword $(MAKEFILE_LIST)) test TEST_FLAGS=-count=1
	$(Q) $(MAKE) -f $(firstword $(MAKEFILE_LIST)) build-ci build-cross-mipsle

verify: ci ## Same as ci (alternative target name for automation).

##@ Run

run-app: ## Placeholder: CLI/router binary (no dev server). Use build-local or cross-build targets.
	@true

##@ Refresh

refresh-data: ## Placeholder: no upstream dataset refresh job in this repo.
	@true

##@ Deploy

deploy-app: ## Placeholder: production deploy is packaging/upload (see pkgs, upload). Not automated here.
	@true

version: ## Print VER (git describe by default; override: make version VER=1.2.3).
	@echo $(VER)

release: all commit push ; @ $(info creating release…) ## Full build + git commit/tag + push (maintainer).
	@echo Released $(TAG)

commit: ; @ $(info committing to git repo) ## Commit release and create annotated tag.
	@echo Committing release $(TAG)
	git commit -am"Release $(TAG)"
	git tag -a $(TAG) -m"Release version $(TAG)"

push: ; $(info $(M) pushing release tags $(TAG) to master…) @ ## Push tags and branch.
	@echo Pushing release $(TAG) to master
	git push --tags
	git push

repo: ; $(info $(M) updating debian repository with version $(TAG)…) @ ## Publish Debian repo via aws.sh (maintainer).
	./aws.sh $(AWS) $(TARGET)_$(VER)_ $(TAG)

upload: pkgs ; $(info $(M) uploading pkgs to test routers…) @ ## scp debs to dev routers (maintainer hosts).
	scp $(TARGET)_$(VER)_mips.deb dev1:/tmp
	scp $(TARGET)_$(VER)_mipsel.deb er-x:/tmp
	scp $(TARGET)_$(VER)_mips.deb ubnt:/tmp

tags: ; @ $(info pushing git tags…) ## git push origin --tags only.
	git push origin --tags

##@ Clean

clean: ; @ $(info cleaning artefacts…) ## Remove build outputs, test artifacts, debs in tree.
	$(GOCLEAN)
	@find . -name "$(EXE).*" -type f \
	-o -name debug -type f \
	-o -name "*.deb" -type f \
	-o -name debug.test -type f \
	-o -name "*.tgz" -type f \
	| xargs rm -f
	@rm -rf test/tests.* test/coverage.*
	@rm -rf /tmp/testBlocklist*
