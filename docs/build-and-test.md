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
| **`make docker-image-e2e`** | Build or rebuild the E2E image (**`Dockerfile.e2e`**: Go + **`dnsmasq`**). |
| **`make test`** | Run **`go test ./...`** inside the dev container. |
| **`make test-e2e`** | Run **`go test ./...`** inside the E2E container as **root** (parity tests that need **`/etc/init.d/dnsmasq`**). |
| **`make build`** | Produce **`dist/update-dnsmasq.mips`** (linux/mips64, ER-Lite class) and **`dist/update-dnsmasq.mipsel`** (linux/mipsle, ER-X class). |
| **`make build-mips64`** | **`dist/update-dnsmasq.mips`** only. |
| **`make build-mipsle`** | **`dist/update-dnsmasq.mipsel`** only. |
| **`make build-e2e`** | **`dist/update-dnsmasq.e2e`** (**linux/amd64** harness binary; not for routers). |
| **`make pkgs`** | Build both architecture-specific **`.deb`** files (and companion **`.tgz`** archives when packaging completes) into **`dist/`**. |
| **`make pkg-mips`** / **`make pkg-mipsel`** | Build one **`.deb`** after the matching **`build-mips*`** binary exists. |
| **`make clean`** | Delete **`dist/`**, stray **`edgeos-adblock_*.deb`** in the repo root, cross-built **`update-dnsmasq.*`** files, and common test artifacts. |
| **`make guard-makefile`** | Run **`scripts/check-makefile-conventions.sh`** (category header order for this Makefile). |

## What gets built

| Artifact | Role |
|----------|------|
| **`dist/update-dnsmasq.mips`** | Static binary for **linux/mips64**. |
| **`dist/update-dnsmasq.mipsel`** | Static binary for **linux/mipsle**. |
| **`dist/update-dnsmasq.e2e`** | Optional **linux/amd64** binary for manual runs inside the E2E image. |
| **`dist/edgeos-adblock_<version>_mips.deb`** / **`_mipsel.deb`** | Router packages; **`<version>`** defaults to the short git commit id (see **`VER`**). |
| **`dist/edgeos-adblock_<version>_*.deb.tgz`** | Optional archive produced alongside each **`.deb`**. |

## Overrides (make variables)

These are the supported knobs for local and CI use. Values are passed into **`docker run`** where applicable.

| Variable | Default | Purpose |
|----------|---------|---------|
| **`DEV_IMAGE`** | **`edgeos-adblock-dev:latest`** | Image name for **`docker build`** / **`docker run`**. |
| **`DOCKERFILE`** | **`Dockerfile.dev`** | Dockerfile path for the dev image. |
| **`E2E_IMAGE`** | **`edgeos-adblock-e2e:latest`** | Image name for **`make docker-image-e2e`** / **`make test-e2e`**. |
| **`DOCKERFILE_E2E`** | **`Dockerfile.e2e`** | Dockerfile path for the E2E image. |
| **`GO_VERSION`** | **`1.26`** (keep aligned with **`go.mod`**) | **`Dockerfile.dev`** and **`Dockerfile.e2e`** build argument. |
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

### End-to-end scope vs EdgeOS

Two different surfaces are easy to conflate:

| Surface | What it validates | Practical automation |
|--------|-------------------|----------------------|
| **Vyatta / package install** | **`.deb`**, templates under **`/opt/vyatta/...`**, **`post-install.sh`** using **`vyatta-cfg-cmd-wrapper`**, **`system task-scheduler`** | **Lab hardware** (EdgeRouter). Ubiquiti does not ship EdgeOS as a generic VM or container image; this path is not reproduced faithfully in Docker. |
| **`update-dnsmasq` runtime** | Config load (**`-f`** or **`config.boot`**), list fetch, dnsmasq include files, reload via init | **`make test-e2e`** (root + **`dnsmasq`** in the E2E image) or the same checks on a router. |

The E2E image is **EdgeOS-shaped for the CLI and dnsmasq only**: Debian Bookworm, **`dnsmasq`** with **`/etc/init.d/dnsmasq`**, and **native linux/amd64** **`go test`** / **`go build`**. It does **not** run Vyatta **`configure`** or install router **`.deb`** packages.

### E2E test target

| Target | Outcome |
|--------|---------|
| **`make docker-image-e2e`** | Build **`edgeos-adblock-e2e:latest`** (override with **`E2E_IMAGE`**) from **`Dockerfile.e2e`**. |
| **`make test-e2e`** | Same as **`make test`**, but inside the E2E image as **root**, so parity tests that require **`/etc/init.d/dnsmasq`** run instead of skipping. Passes **`-count=1`** so the Go test cache does not replay **`make test`** results from a non-root environment. Uses **`TEST_FLAGS`**, **`TEST_TIMEOUT`**, and **`GO_CACHE_VOLUME`** like **`make test`**. |
| **`make build-e2e`** | Optional **linux/amd64** binary **`dist/update-dnsmasq.e2e`** for manual smoke runs in the E2E container (not a router build). |

**`make test-e2e`** needs outbound HTTPS (for example **`ChkWeb`** and live list fixtures). In restricted networks, expect failures unless you adjust tests or provide a proxy.

The **`internal/dnsmasq`** tests include **`TestDnsmasqServesBlockedLookups`**, which starts a real **dnsmasq** bound to **127.0.0.1** on an ephemeral UDP port, loads merged **`address=/…/0.0.0.0`** lines from **`internal/testdata/etc/dnsmasq.d/`**, and checks that a random subset of those names resolve to **0.0.0.0** via **`LookupIPAddr`**. It runs when the **dnsmasq** binary is on **`PATH`** (the E2E image) and is skipped in the dev image.

## Runtime container image

The repository ships **`Dockerfile.dev`** (development) and **`Dockerfile.e2e`** (root + dnsmasq for parity tests). There is no separate minimal “runtime-only” Dockerfile in-tree; use **`Dockerfile.e2e`** or derive a slim image from Debian if you need a smaller deployable test runner.

## Change expectations

From a clean clone with only Docker and make installed, **`make guard-makefile`**, **`make test`**, **`make build`**, and **`make pkgs`** should succeed after **`make docker-image`** (or any target that triggers the first image build). **`make test-e2e`** additionally requires outbound HTTPS to complete parity tests that contact the live network.
