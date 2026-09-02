# Makefile conventions

The Makefile is the contributor interface for this repository. `make help`
lists public targets and is the default target.

## Target groups

Public targets are grouped in this order:

1. **Setup** — images and repository checks
2. **Build** — binaries and Debian packages
3. **Quality** — lint checks
4. **Test** — automated tests
5. **Clean** — generated-file cleanup

Each public target has a short `##` description so it appears in `make help`.
Targets that create files or start containers should say so in that
description.

## Adding a target

Choose the existing group that matches the target, keep the target name
imperative and specific, and add it to `.PHONY` when it is not a file-producing
target. Keep Go and packaging commands inside the Docker image so a fresh clone
needs only Docker and make.

Run these checks before submitting a Makefile change:

```bash
make guard-makefile
make lint
```

The guard is implemented by
[`scripts/check-makefile-conventions.sh`](../scripts/check-makefile-conventions.sh)
and is also run by CI.
