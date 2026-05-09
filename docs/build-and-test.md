# Build and test

Use the Makefile as the canonical interface (`make help`). **There is no `vendor/` directory** — modules are fetched into the cache (`go mod download`), or you build inside Docker for a fully pinned environment. The installable command is **`./cmd/update-dnsmasq`** (import path **`github.com/jonmeacham/edgeos-adblock/cmd/update-dnsmasq`**, `package main`).

## Docker (preferred for Linux binaries)

From the repo root:

```bash
docker build -t edgeos-adblock:local .
# or
make docker-build
```

The **`Dockerfile`** uses **`golang:<version>-bookworm`** to **`go build`** (modules downloaded inside the layer cache), then installs the binary into **`debian:bookworm-slim`** with CA certificates. Tests are **not** run in the image by default (they expect EdgeOS-style paths and syslog); use **`make ci`** on a dev machine or your CI runner for **`go test`**. Adjust **`ARG GO_VERSION`** to match **`go.mod`** / **`toolchain`**.

Override image name:

```bash
docker build -t my-registry/edgeos-adblock:latest .
```

## Everyday development (host)

| Command | Purpose |
|--------|---------|
| `make install` | `go mod download`, `verify`, `tidy`; install `golangci-lint` into `$(go env GOPATH)/bin` if missing. |
| `make test` | `go test ./...` (135s timeout). Use `make test TEST_FLAGS=-count=1` to bypass test cache. |
| `make build-local` | Build `update-dnsmasq` from **`cmd/update-dnsmasq`** for your OS/arch (`-mod=readonly`). Does **not** run `go generate`; keep committed stringer output in sync when enums change. |
| `make fmt` / `make format-check` | Apply `gofmt` or fail if formatting drifts. |

## Merge / pipeline gate (host)

Run **`make ci`** (same as **`make verify`**):

1. `guard-makefile`
2. `install`
3. `format-check`, `vet`, `lint`
4. `test` with **`TEST_FLAGS=-count=1`**
5. `build-ci` — host compile check → `$(DISTDIR)/edgeos-adblock.ci` (default `/tmp`)
6. `build-cross-mipsle` — linux/mipsle cross-compile (no `go generate`)

For **reproducible** Linux builds and tests in CI, run **`docker build`** (or your own image that mirrors the Dockerfile) instead of relying on the host toolchain.

## Targets that modify files

- **`make check`** / **`make tests`** run **`fmt`** then **`install`** (module tidy/download) then tests — they can rewrite sources. Prefer **`make test`** for read-only validation.

## Coverage and extras

- **`make test-coverage`** — merged HTML/XML coverage (optional gocov tools under `$(BIN)/`).
- **`make generate`** — stringer via `go generate`; needs a matching **`golang.org/x/tools`** in the module graph (use the same Go version as **`go.mod` `toolchain`**).
- **`make mipsle`**, **`make build`**, **`make pkgs`** — full release / `.deb` paths (still use **make** for EdgeOS packages; use **Docker** when you want a clean module-only build for Linux).

## Router install (EdgeOS)

Install the **`edgeos-adblock`** `.deb` for your router CPU, or use **`make_deb`** / your release pipeline. Vyatta templates live under **`.payload/`**; **`post-install.sh`** seeds blacklist sources and schedules **`update-dnsmasq`**. For UniFi **`config.gateway.json`** provisioning, see the sample at the repo root.

## Contract

Anything you remove from the tree should still allow **`make ci`** (host) and/or a successful **`docker build`** (container), depending on which path you support in your infrastructure.
