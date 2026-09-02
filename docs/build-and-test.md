# Build, test, and package

The Makefile runs formatting checks, `go vet`, tests, cross-compilation, and
Debian packaging directly in the upstream Go 1.26.3 Bookworm image pinned by
digest. A contributor needs Docker and GNU make on the host. No custom image or
additional package installation is involved; Debian packages are built with
the base image's `dpkg-deb`.

## Targets

Run `make help` for the authoritative target list.

| Target | Result |
| --- | --- |
| `check` | Run the complete CI verification and package build. |
| `guard-makefile` | Validate the Makefile contributor interface. |
| `lint` | Check `gofmt` output and run `go vet ./...`. |
| `test` | Run `go test ./...`. |
| `build` | Build both EdgeOS architectures. |
| `build-mips64` | Build `dist/edgeos-adblock.mips`. |
| `build-mipsle` | Build `dist/edgeos-adblock.mipsel`. |
| `pkgs` | Build both Debian packages. |
| `pkg-mips` | Build the mips Debian package. |
| `pkg-mipsel` | Build the mipsel Debian package. |
| `clean` | Remove generated build and test artifacts. |

The package version defaults to `0.0.0+git.<short-revision>`. Package files
are written to `dist/` as `edgeos-adblock_<version>_<architecture>.deb`.

## Variables

| Variable | Default | Purpose |
| --- | --- | --- |
| `GO_IMAGE` | Go 1.26.3 Bookworm digest | Upstream build image; keep aligned with `go.mod`. |
| `GO_CACHE_VOLUME` | `edgeos-adblock-go-cache` | Docker Go cache volume. |
| `VER` | `0.0.0+git.<revision>` | Package and binary version. |
| `GIT` | short Git revision | Embedded source revision. |
| `GOMIPS64` | empty | Optional mips64 floating-point ABI. |
| `MIPSLE_GOMIPS` | `softfloat` | mipsle floating-point ABI. |
| `TEST_FLAGS` | empty | Additional `go test` arguments. |
| `TEST_TIMEOUT` | `135s` | Test timeout. |
| `VERBOSE` | `0` | Set to `1` to display recipe commands. |

Examples:

```bash
TEST_FLAGS='-count=1' make test
MIPSLE_GOMIPS=hardfloat make build-mipsle
VER=1.0.0 make pkgs
```

## Package contents

Packaging stages the selected static binary at
`.payload/config/scripts/edgeos-adblock`, then invokes `make_deb`, which uses
`dpkg-deb` directly. The resulting package contains the updater, Vyatta
templates, and lifecycle hooks only. The staged binary and all output under
`dist/` are ignored by Git.
