# Makefile conventions

This is the team-wide standard for `Makefile`s in repos we maintain. The Makefile is the canonical contributor interface; `make help` is the exhaustive target catalog. Reach for direct commands (`go`, `golangci-lint`) only when targeting a specific component or debugging.

The standard is portable — copy this doc and the guard script into a new project to enforce it.

## Skeleton

```make
.PHONY: \
    help \
    install \
    build test lint typecheck coverage ci \
    build-app test-app run-app deploy-app clean

.DEFAULT_GOAL := help

help: ## Show this target catalog.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n"} /^##@/ { printf "\n%s\n", substr($$0, 5) } /^[A-Za-z0-9_.-]+:.*##/ { printf "  %-38s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Setup
install: ## Install project dependencies.
	...

##@ Build
build: ## Build all workspaces / packages.
	...

##@ Test
test: ## Run baseline tests.
	...

##@ Run
run-app: ## Start the app locally.
	...

##@ Deploy
deploy-app: ## Deploy the app.
	...

##@ Clean
clean: ## Remove local generated artefacts.
	...
```

## Categories (canonical order)

`##@` headers must appear in this order. The guard script enforces it.

1. **Setup** — install / sync / bootstrap / link binaries
2. **Build** — compile, generate, rebuild artefacts (`build-*`)
3. **Quality** — lint / format-check / typecheck / validate-schemas (`lint-*`, `format-*`, `typecheck-*`, `validate-*`)
4. **Test** — unit, integration, E2E, component lanes (`test-*`)
5. **Coverage** — coverage gates (`coverage-*`)
6. **CI** — `ci` and `ci-*` aggregate gates suitable for pre-push or pipeline use
7. **Run** — local execution / dev servers (`run-*`)
8. **Refresh** — external data / API refresh workflows (`refresh-*`)
9. **Deploy** — infrastructure or production-affecting deployments (`deploy-*`)
10. **Clean** — local destructive cleanup (`clean-*`)

If a project has nothing to do in a category, include it anyway with at least one placeholder target — the guard asserts the full ordered list.

## Naming

| Prefix | Use for |
| --- | --- |
| `build-*` | compile, generate, rebuild artefacts |
| `test-*` | unit, integration, E2E, component-specific tests |
| `run-*` | local execution / dev servers |
| `refresh-*` | external data / API refresh workflows |
| `deploy-*` | infrastructure or production-affecting deployments |
| `coverage-*` | coverage gates |
| `lint-*`, `format-*`, `typecheck-*`, `validate-*`, `ci-*` | quality lanes |
| `clean-*` | local destructive cleanup |

## Documentation rules

- Every public `.PHONY` target has inline `## help text` after the recipe colon. If it's not in `make help`, it's effectively a private target — make it private (don't list in `.PHONY`).
- Group public targets under `##@ Category` headers in the canonical order above.
- Keep target descriptions short, imperative, and explicit about side effects.
- Repo docs explain workflows; they don't duplicate the full target catalog.
- Point contributors to `make help` for the complete list.

## Guard

`scripts/check-makefile-conventions.sh` validates `.DEFAULT_GOAL := help` and the ordered `##@` headers expected for that repository. In **edgeos-adblock** the guard expects **Setup → Build → Test → Clean** (Docker-only slim Makefile). Run **`make guard-makefile`**.

## Quick adoption checklist

1. Copy `scripts/check-makefile-conventions.sh` (or adapt category list inside it).
2. Rewrite the `Makefile` per the skeleton and categories above (or your chosen subset, and align the guard script).
3. Run the guard; iterate until green.
4. Invoke your CI entrypoint (`make test`, `make ci`, etc.) from your runner or hooks.
5. Reference this doc from the project's contributor docs.
