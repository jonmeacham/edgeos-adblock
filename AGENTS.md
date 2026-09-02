# EdgeOS repository guidance

Treat the Makefile as the contributor interface and preserve the category and
help conventions documented in `docs/makefile-conventions.md`.

EdgeOS configuration is transactional. Package scripts must use the Vyatta
configuration wrapper (or the `configure`, `set`, `commit`, `save`, `exit`
workflow), check commit failures, and never edit `/config/config.boot` or files
under `/opt/vyatta/config/` directly.

Router scripts use `#!/bin/vbash`, source
`/opt/vyatta/etc/functions/script-template`, run non-interactively with the
`vyattacfg` group, and remain safe to run more than once.
