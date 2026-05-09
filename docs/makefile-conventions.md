# Makefile conventions

This document describes how we structure **`Makefile`**s in repositories we maintain. The Makefile is the main contributor interface; **`make help`** is the catalog of supported targets. Use ad hoc **`go`**, **`docker`**, or other commands only when you are debugging a specific step.

When you write or edit **repository documentation**, prefer stable behavior (how to run, configure, cache, override, and clean) over transient details (exact log lines, one-off toolchain warnings, or internal flags) unless those details are part of an intentional public contract.

## This repository

**edgeos-adblock** uses a **Docker-only** Makefile: Go runs inside **`Dockerfile.dev`**, and the host needs Docker and make only. Public **`##@` headers** are intentionally minimal:

1. **Setup** — dev image and convention checks  
2. **Build** — cross-compile and **`.deb`** packaging  
3. **Test** — **`go test`** in the dev image  
4. **Clean** — remove **`dist/`**, packages, and common test outputs  

**`scripts/check-makefile-conventions.sh`** asserts **`.DEFAULT_GOAL := help`** and that those four headers appear in that order. Run **`make guard-makefile`** before you push Makefile changes.

For what each target does, which variables you can set, and where artifacts go, see **[build-and-test.md](build-and-test.md)** and **`make help`**.

## Full standard (other repositories)

Larger projects may use more **`##@` categories** in this fixed order:

1. **Setup** — install, sync, bootstrap  
2. **Build** — compile and generated artifacts (**`build-*`**)  
3. **Quality** — lint, format-check, typecheck (**`lint-*`**, **`format-*`**, …)  
4. **Test** — unit and integration lanes (**`test-*`**)  
5. **Coverage** — coverage gates (**`coverage-*`**)  
6. **CI** — aggregates for pipelines (**`ci`**, **`ci-*`**)  
7. **Run** — local dev servers (**`run-*`**)  
8. **Refresh** — data or API refresh (**`refresh-*`**)  
9. **Deploy** — production or infra (**`deploy-*`**)  
10. **Clean** — destructive local cleanup (**`clean-*`**)

If a category has nothing to do yet, keep a placeholder target so the guard script and **`make help`** stay consistent.

### Naming

| Prefix | Use for |
| --- | --- |
| **`build-*`** | Compile, generate, rebuild artifacts |
| **`test-*`** | Tests |
| **`run-*`** | Local execution |
| **`refresh-*`** | External data refresh |
| **`deploy-*`** | Deployments |
| **`coverage-*`** | Coverage gates |
| **`lint-*`**, **`format-*`**, **`typecheck-*`**, **`validate-*`**, **`ci-*`** | Quality and CI |
| **`clean-*`** | Local cleanup |

### Documentation rules for Makefiles

- Every public **`.PHONY`** target has **`##` help text** after the recipe colon. If it is not in **`make help`**, treat it as private (omit from **`.PHONY`** or document that it is internal).  
- Group public targets under **`##@ Category`** headers in the canonical order for that repo.  
- Keep descriptions short, imperative, and explicit about important side effects (writes **`dist/`**, starts containers, etc.).  
- Do not duplicate the full target list in prose; point readers to **`make help`**.

## Guard script

Copy **`scripts/check-makefile-conventions.sh`** into a new project and adjust the **`expected=(...)`** array to match the **`##@` headers** you use. Run it from **`make guard-makefile`** or CI.

## Adoption checklist

1. Copy or adapt **`scripts/check-makefile-conventions.sh`**.  
2. Add **`help`** as the default goal and ordered **`##@`** sections.  
3. Run the guard until it passes.  
4. Wire **`make test`** (or **`make ci`**) into your CI or pre-push workflow.  
5. Link this doc from the project README for contributors.
