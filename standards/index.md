# Standards Index

Standing instructions for all development projects. Follow each one where it
applies.

- `go/` — Go standards.
  - `writing.md` — project layout, application identity, logging, errors,
    concurrency, security, and style.
  - `tests.md` — the `make test` contract and how Go test suites are built.
- `m5/` — M5Stack device notes.
  - `paper-color.md` — driving the M5Paper Color e-paper display.
- `makefile/` — standard Makefile targets and their contracts.
  - `makefile.md` — `make`, `test`, `build`, `clean`, and `install`.
- `services/` — `install` and `uninstall` subcommands for services and
  daemons, one file per platform.
  - `linux.md` — systemd units.
  - `macos.md` — launchd plists.
  - `windows.md` — Service Control Manager registration.
- `ux/` — UX guidance principles.
  - `ux.md` — sourcing rules for UX advice and review.
