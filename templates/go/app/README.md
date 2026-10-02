# `app/` — application identity template

Holds the application's name, tagline, copyright and version. Copy it into new
Go projects and fill in the constants; do not change anything else. The rules
are in `standards/go/writing.md` under "Versioning and Copyright"; this
directory is the source of truth for the code.

## Install

1. Copy `app.go` and `app_test.go` into `app/` at the repository root, not
   under `pkg/`.
2. Fill in the copyright header in both files (see
   `templates/copyright/header.md`).
3. Set `name`, `tagLine`, `copyright` and `version` in `app.go`. New projects
   start at `0.1.0`.
4. Replace `<module>` in the ldflags comment with the module path.
5. Add the ldflags below to the Makefile.

## Makefile

```make
GIT_COMMIT=$(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION=$(shell go version | awk '{print $$3}')
# UTC so it always increases; := so build-all stamps one number on every target.
BUILD_NUMBER:=$(shell date -u +%Y%m%d%H%M%S)
APP_PKG=<module>/app
LDFLAGS=-ldflags "-X $(APP_PKG).gitCommit=$(GIT_COMMIT) -X $(APP_PKG).buildTime=$(BUILD_TIME) -X $(APP_PKG).goVersion=$(GO_VERSION) -X $(APP_PKG).buildNumber=$(BUILD_NUMBER) -s -w"
```

Never inject the version (e.g. from `git describe`); it belongs to the source.

## Use

| Call | Returns | For |
| --- | --- | --- |
| `app.Name()` | `ExampleApp` | product name |
| `app.TagLine()` | `Does the thing` | banner, `--version` |
| `app.Copyright()` | `Copyright (c) …` | banner, `--version` |
| `app.Version()` | `0.1.0+1a2b3c4d [20260902155301]` | display: banner, `--version`, logs |
| `app.SemVer()` | `0.1.0` | protocol handshakes, comparisons |
| `app.Build()` | `20260902155301` | ordering builds; `""` if unstamped |
| `app.BuildDate()` | `20260902` | int-only fields; `0` if unstamped |
| `app.BuildInfo()` | timestamp, toolchain | `--version` detail |

Always use the accessors.

## Bumping the version

Edit `version` in `app.go`. Don't tag as a side effect; tagging is part of the
release process.
