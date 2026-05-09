# Build, test, and package

This document is for people who change the code or produce installable artifacts. It describes what you need on your machine, what **`make`** targets do, where outputs land, and which variables you can override without editing the Makefile.

## Prerequisites

- **Docker** (for builds, tests, and `.deb` creation)
- **GNU make**

You do not need a host-installed Go toolchain.

## Dev image

The default image name is **`edgeos-adblock-dev:latest`** (override with **`DEV_IMAGE`**). The image is defined in **`Dockerfile.dev`**. The first **`make test`**, **`make build`**, or **`make pkgs`** run builds it if it is missing.

## Typical workflow

```bash
make test
make build
make pkgs
```

Packages and binaries appear under **`dist/`** (ignored by git). Remove them with **`make clean`**.

## Targets

Use **`make help`** for the authoritative list with short descriptions. The table below summarizes outcomes.

| Target | Outcome |
|--------|---------|
| **`make docker-image`** | Build or rebuild the dev image only. |
| **`make test`** | Run **`go test ./...`** inside the dev container. |
| **`make build`** | Produce **`dist/update-dnsmasq.mips`** (linux/mips64, ER-Lite class) and **`dist/update-dnsmasq.mipsel`** (linux/mipsle, ER-X class). |
| **`make build-mips64`** | **`dist/update-dnsmasq.mips`** only. |
| **`make build-mipsle`** | **`dist/update-dnsmasq.mipsel`** only. |
| **`make pkgs`** | Build both architecture-specific **`.deb`** files (and companion **`.tgz`** archives when packaging completes) into **`dist/`**. |
| **`make pkg-mips`** / **`make pkg-mipsel`** | Build one **`.deb`** after the matching **`build-mips*`** binary exists. |
| **`make clean`** | Delete **`dist/`**, stray **`edgeos-adblock_*.deb`** in the repo root, cross-built **`update-dnsmasq.*`** files, and common test artifacts. |
| **`make guard-makefile`** | Run **`scripts/check-makefile-conventions.sh`** (category header order for this Makefile). |

## What gets built

| Artifact | Role |
|----------|------|
| **`dist/update-dnsmasq.mips`** | Static binary for **linux/mips64**. |
| **`dist/update-dnsmasq.mipsel`** | Static binary for **linux/mipsle**. |
| **`dist/edgeos-adblock_<version>_mips.deb`** / **`_mipsel.deb`** | Router packages; **`<version>`** defaults to the short git commit id (see **`VER`**). |
| **`dist/edgeos-adblock_<version>_*.deb.tgz`** | Optional archive produced alongside each **`.deb`**. |

## Overrides (make variables)

These are the supported knobs for local and CI use. Values are passed into **`docker run`** where applicable.

| Variable | Default | Purpose |
|----------|---------|---------|
| **`DEV_IMAGE`** | **`edgeos-adblock-dev:latest`** | Image name for **`docker build`** / **`docker run`**. |
| **`DOCKERFILE`** | **`Dockerfile.dev`** | Dockerfile path for the dev image. |
| **`GO_VERSION`** | **`1.26`** (keep aligned with **`go.mod`**) | **`Dockerfile.dev`** build argument. |
| **`GO_CACHE_VOLUME`** | **`edgeos-adblock-go-cache`** | Docker volume name mounted at **`/cache`** for Go module and build caches. |
| **`VER`** | Short **`git`** commit id | Embedded **`main.version`** and **`.deb`** / **`.tgz`** filenames. Set explicitly for a release label without changing git. |
| **`GIT`** | Short **`git`** commit id | Embedded **`main.githash`**; defaults to match **`VER`**. |
| **`GOMIPS64`** | *(empty)* | Optional **linux/mips64** floating-point ABI (**`softfloat`** if you need it). |
| **`MIPSLE_GOMIPS`** | **`softfloat`** | **linux/mipsle** **`GOMIPS`**; use **`hardfloat`** or empty string when your toolchain and device require it. |
| **`TEST_FLAGS`** | *(empty)* | Extra arguments to **`go test`** (for example **`-count=1`**). |
| **`TEST_TIMEOUT`** | **`135s`** | **`go test`** timeout. |
| **`VERBOSE`** | **`0`** | Set to **`1`** to print recipe lines instead of **`@`**-quiet rules. |

## Go build cache

The Makefile mounts a named Docker volume so module download and compile caches survive across **`docker run`** invocations. Use a different volume per clone or CI job with **`GO_CACHE_VOLUME`**. To force a completely cold build, remove the volume (for example **`docker volume rm edgeos-adblock-go-cache`**) and run **`make build`** again.

## Packages

**`make pkgs`** stages the correct **`update-dnsmasq`** binary into **`.payload/config/scripts/`**, runs **`make_deb`** inside the dev container, and moves **`.deb`** / **`.tgz`** results into **`dist/`**. The **`.deb`** installs Vyatta templates under **`/opt/vyatta/...`** and scripts under **`/config/scripts/`**; first install behavior is defined by **`.payload/post-install.sh`** (Vyatta **`configure`** session that enables **`service dns forwarding blocklist`** sources, excludes, cron, and related settings).

## Tests

**`make test`** runs the full Go test suite in Docker. Some tests expect **root** and an EdgeOS-style **`/etc/init.d/dnsmasq`**; those are skipped when that environment is not present (typical laptops and the standard dev container).

## Runtime container image

For a minimal runtime image (not the dev image), build from the repository **`Dockerfile`**:

```bash
docker build -t edgeos-adblock:local .
```

Keep the Dockerfile’s Go version argument in step with **`go.mod`** when you upgrade the language version.

## Change expectations

From a clean clone with only Docker and make installed, **`make guard-makefile`**, **`make test`**, **`make build`**, and **`make pkgs`** should succeed after **`make docker-image`** (or any target that triggers the first image build).
