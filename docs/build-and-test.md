# Build and test

The **Makefile** is the canonical interface (`make help`). **Go runs only inside Docker** — the host needs **Docker** and **GNU make** (no local Go toolchain).

## Dev image (`Dockerfile.dev`)

`make test` and `make build` use **`edgeos-adblock-dev:latest`** (override with **`DEV_IMAGE=...`**). The image is built from **`Dockerfile.dev`**; keep **`GO_VERSION`** in sync with **`go.mod`** / **`toolchain`**.

The Makefile mounts a **Docker named volume** (default **`edgeos-adblock-go-cache`**) for **`GOMODCACHE`** and **`GOCACHE`**, so incremental builds reuse the Go module and build caches across container runs. Set **`GO_CACHE_VOLUME`** to use a different volume name; **`docker volume rm edgeos-adblock-go-cache`** clears the cache for a fully cold build.

| Command | Purpose |
|--------|---------|
| `make docker-image` | Build or rebuild the dev image. |
| `make test` | `go mod download` then `go test ./...` inside the container (`TEST_FLAGS`, `TEST_TIMEOUT` optional). |
| `make build` | Builds **both** router binaries in Docker: **`dist/update-dnsmasq.mips`** (`linux/mips64`, ER-Lite class) and **`dist/update-dnsmasq.mipsel`** (`linux/mipsle`, ER-X class). Same names as the legacy packaging flow. |
| `make build-mips64` | **`linux/mips64`** only → **`dist/update-dnsmasq.mips`**. Optional **`GOMIPS64=softfloat`**. |
| `make build-mipsle` | **`linux/mipsle`** only → **`dist/update-dnsmasq.mipsel`**. Default **`MIPSLE_GOMIPS=softfloat`**; override with **`MIPSLE_GOMIPS=hardfloat`** or empty. |
| `make pkgs` | Builds **both** `.deb` packages (`make pkg-mips` + `make pkg-mipsel`). Installs the matching **`update-dnsmasq`** under **`/config/scripts/`**, ships Vyatta templates under **`/opt/vyatta/...`**, and on **first install** runs **`.payload/post-install.sh`** (Vyatta `configure` session) to enable **`service dns forwarding blocklist`** sources, excludes, cron, etc. Output: **`dist/edgeos-adblock_<ver>_mips.deb`** and **`_mipsel.deb`** (plus optional **`.tgz`** archives). |
| `make pkg-mips` / `make pkg-mipsel` | Build a single-arch `.deb` after the matching **`build-mips*`** binary exists. |
| `make clean` | Remove **`dist/`**, cross-build `update-dnsmasq.*` artefacts, and common test outputs. |
| `make guard-makefile` | Run **`scripts/check-makefile-conventions.sh`**. |

Example:

```bash
make test TEST_FLAGS=-count=1
make build
make pkgs   # .deb files in dist/
```

Some integration tests require **root** and an EdgeOS-style **`/etc/init.d/dnsmasq`**; those are skipped automatically in the dev container and on typical workstations.

## Runtime image (`Dockerfile`)

Shippable OCI image (Debian slim runtime, `CGO_ENABLED=0` build in the multi-stage **`Dockerfile`**):

```bash
docker build -t edgeos-adblock:local .
```

Adjust **`ARG GO_VERSION`** in **`Dockerfile`** to match **`go.mod`**.

## Contract

Changes should keep **`make guard-makefile`**, **`make test`**, **`make build`**, and **`make pkgs`** succeeding from a clean clone with only Docker + make installed.
