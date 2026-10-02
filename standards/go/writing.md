# Go Coding Standards

## Project Structure

Packages live at the root of the source tree, one concept per package. Three
directories have a reserved meaning:

| Directory | Holds |
| --- | --- |
| `app/` | Application identity: name, tagline, copyright, version. See Versioning and Copyright. |
| `internal/` | Private implementation. The compiler refuses imports from outside the module, so this is a guarantee rather than a convention. Anything not deliberately part of a public API belongs here. |
| `cmd/` | Additional executables. A single-binary application keeps `main.go` at the root; use `cmd/<name>/` for further binaries or a repository mixing importable packages with commands. |

```
project-root/
  main.go            single binary: keep it here
  app/               application identity
  config/            root packages, one concept each
  server/
  handlers/          nest ONLY to group a real family
    webhook/
    admin/
  internal/          private implementation, compiler-enforced
  cmd/               additional binaries
```

Nest only to group a genuine family: `handlers/webhook`, `handlers/admin`,
`channels/slack`, `channels/telegram`. A parent directory that exists only as a
container is not a family.

### Do not use `pkg/`

`pkg/` is not a Go convention. It appears in neither the standard library
(`net/http`, not `pkg/net/http`) nor Go's own module-layout guidance, which
documents `internal/` and `cmd/` and nothing else.

We do not use pkg/ for two reasons:

- It adds a path segment carrying no information. Every package in a module is
  a package; saying so in the import path is like a directory named `files/`.
- It is routinely used where `internal/` was meant. `pkg/` *looks* like it
  signals "importable, deliberate API" against a private remainder, but it
  enforces nothing, while `internal/` enforces exactly that.

In a repository that already has `pkg/`:

- **Create new packages in the right place**: the root, `internal/`, or `cmd/`.
  Never add to `pkg/`, and never mirror the existing layout "for consistency".
- **Tell the user that `pkg/` does not comply with these standards**, so the
  decision to migrate is theirs to make deliberately.
- **Do not move anything out of `pkg/` without explicit instruction.** A
  module-wide rename rewrites every import path in the tree: it collides with
  every open branch, buries unrelated review, and is not a side effect any
  other task gets to have. Wait to be asked.

### Other structure rules

- Avoid cyclic dependencies.
- Keep `main.go` small: wire dependencies, run the app, handle graceful exit,
  and report fatal errors. Business logic and config loading belong in
  packages.
- Ask before preserving legacy behaviour or adding backward compatibility
  during development.

## Versioning and Copyright

Every application defines its identity once, in a root-level `app/` package,
and reaches it only through accessors. Nothing else names or versions the
application.

The package is a template, kept as real code rather than a snippet in this
document: `templates/go/app/` in this repository. Copy the directory to the
project root, change the four constants, and replace `<module>` in the ldflags
comment. Its `README.md` covers installation and the Makefile block.

Keeping the code in one place and describing it here is deliberate. A copy of
the source inlined in this document is a second definition that will drift
from the first, which is the exact defect this section exists to prevent.

```
templates/go/app/
  app.go        the package: four constants, four injected vars, eight accessors
  app_test.go   pins the properties that regress silently
  README.md     install steps, Makefile ldflags, usage table
```

### The constants

| Constant | Meaning |
| --- | --- |
| `name` | Product name, typically the repository/module name. |
| `tagLine` | One-line description, for the banner and `--version`. |
| `copyright` | Matches the year and holder in the project `LICENSE`. |
| `version` | Semantic version, bare. New projects start at `0.1.0`. |

### The accessors

| Accessor | Returns | Use for |
| --- | --- | --- |
| `Name()` | `ExampleApp` | anywhere the product is named |
| `TagLine()` | `Does the thing` | banner, `--version` |
| `Copyright()` | `Copyright (c) …` | banner, `--version` |
| `Version()` | `0.1.0+1a2b3c4d [20260902155301]` | **display**: banner, `--version`, logs, diagnostics, prompts, APIs meant to be read |
| `SemVer()` | `0.1.0` | **parseable**: protocol handshakes (MCP `serverInfo`, ACP `Implementation`, device handshakes), version comparisons |
| `Build()` | `20260902155301` | **order**: UTC link time as `yyyymmddhhmmss`, or `""` if unstamped |
| `BuildDate()` | `20260902` | **int day**: YYYYMMDD for an existing int-typed protocol or store; `0` if unstamped |
| `BuildInfo()` | timestamp, toolchain | `--version` detail lines |

The git commit is an 8-character short SHA. The build number is the UTC time
the binary was linked (`date -u +%Y%m%d%H%M%S`). It exists because a commit
hash identifies source exactly but has no order: a rebuild of the same commit
is otherwise indistinguishable. A timestamp rather than a counter because a
counter needs stored state. Seconds rather than minutes because two rebuilds
inside one minute is ordinary. UTC because local time runs backwards an hour
twice a year. Stamp it with `:=` in Make so one `make build-all` writes the
same number into every platform binary.

The 14-digit stamp does not fit in a 32-bit `int`. `Build()` is a string for
that reason. `BuildDate()` is the first eight digits as an `int` (YYYYMMDD),
which is safely below `math.MaxInt32` on every platform Go supports. Use it
only when an existing protocol or store already carries build as an `int` and
cannot take the full stamp; it is enough to know how old a build is, not to
tell two builds on the same UTC day apart.

`SemVer` is named for *why* you reach for it rather than for its shape.
"Short" invites someone shortening a log line; `SemVer` says "this is the one
you compare".

No `Get` prefix. Effective Go: *"it's neither idiomatic nor necessary to put
Get into the getter's name."*

### Why everything is unexported

An exported constant sitting beside an accessor gives two ways to read the
same value, and call sites will use both. This has happened: a binary rendered
its own version two different ways depending on which symbol each site reached
for.

`go tool link -X` sets package-level string vars by symbol name and does not
care about case, so the injected metadata can be unexported too. Build tooling
that needs the version reads the `version = "..."` line out of the source.

Do not group the identity into a struct. `-X` cannot write a struct field and
**fails silently with exit code 0**, so every binary would ship a blank commit
from a green build.

### Format rules

**Never derive the version from the build environment.** Injecting
`git describe --tags` over the constant produces a binary whose version depends
on the machine that compiled it and silently disagrees with its own source.
Stamp only the commit, timestamp, toolchain, and build number.

**The release and the commit are one unbroken token.** `0.1.0 (git: 1a2b3c4d)`
gets pasted into a bug report as `0.1.0`: the space reads as the end of the
value and the parenthetical as an aside, so the half identifying the exact
source is the half that gets dropped. The build number sits after a space in
brackets, `0.1.0+1a2b3c4d [20260902155301]`, so truncating there still leaves
the commit attached.

**Attach the commit with `+`, never `-`.** SemVer 2.0.0 gives the two
separators different meanings, and only one of them is what we mean:

- `+` introduces **build metadata**, which item 10 requires be IGNORED when
  determining precedence. `0.1.0+1a2b3c4d` compares **equal** to `0.1.0`: it
  is that release, built from that commit.
- `-` introduces a **pre-release** identifier. Item 11: a pre-release version
  has lower precedence than the normal version. `0.1.0-1a2b3c4d` compares
  **lower** than `0.1.0`, claiming to be something that came *before* the
  release rather than an instance of it.

The spec standardises the slot, not its contents: it does not say `+` means
"git commit" (the examples include timestamps and build numbers), so label the
contents if the project uses something less self-evident than a SHA.

Two places `+` is genuinely awkward: it is invalid in a Docker image tag
(`[A-Za-z0-9_.-]` only) and decodes as a space in a URL query string. If a
project pushes images or embeds versions in URLs, use the commit alone in
those contexts rather than degrading the version format everywhere.

**If an existing application does not already follow this** (identity outside
`app/`, exported constants, a version injected at build time, `Get`-prefixed
or differently-named accessors), **ask before changing it.** Build scripts,
release tooling, deployment checks and protocol clients may all read these
names and formats, and none of them are visible from inside the module.

### File header

Begin every Go source file with the copyright header block from
`templates/copyright/header.md`, with `[YEAR]` and `[COMPANY]` filled in:

```go
/******************************************************************************
 * Copyright (c) [YEAR] [COMPANY]                                             *
 * Please see LICENSE file for details.                                       *
 ******************************************************************************/
```

Use a year range (e.g. `2025-2026`) once a file spans multiple years.

### Startup banner

Applications (services, servers, anything with a lifecycle) display their
name, version, and copyright on startup, built from the `app` accessors.
Small utilities do not; a tool that is piped or scripted stays quiet and
reports its identity only through `--version`.

```go
fmt.Fprintf(os.Stderr, "%s %s\n%s\n", app.Name(), app.Version(), app.Copyright())
```

```
ExampleApp 0.1.0+1a2b3c4d [20260902155301]
Copyright (c) 2026 [COMPANY]
```

- **Interactive applications:** write the banner to **stderr** so
  machine-readable stdout (JSON, CSV, tables) stays clean.
- **Long-running services:** log the banner at `Info` as the first startup
  line.

## Constructors and Dependencies

- Use `New()` constructors and the go functional options pattern for packages that manage
  state, dependencies, configuration, clients, services, or long-lived components.
- Do not add constructors to simple utility or pure-function packages unless useful.
- Verify libraries/frameworks before using them.
- Prefer the standard library and existing project dependencies.
- Ask before adding large frameworks, code generators, services, or public-API-affecting dependencies.
- Do not introduce new external dependencies without user approval.

## Logging

- Use the project's existing logger. If there is none and the project
  instructions file (`CLAUDE.md` or `AGENTS.md`) records no logging decision,
  ask the user which logger to use and record the answer there.
- Define one logger interface (or alias the library's) in a package every
  other package can import without a cycle.
- Instantiate the logger in `main()` and pass it through constructors/options.
- Packages should accept loggers via a `WithLogger(logger)` option.
- Never use package-level loggers unless explicitly approved.
- Never log secrets, credentials, tokens, or sensitive user data.

Levels (use the nearest the chosen logger provides):

- `Debug`: diagnostic detail for development/troubleshooting.
- `Info`: normal flow and successful operations.
- `Notice`: security events and important operational notifications.
- `Warning`: recoverable issues or suspicious conditions.
- `Error`: unexpected or unrecoverable failures.

## Errors

- Always handle errors explicitly.
- Return errors as the last return value in the idiomatic go way.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Error strings are lower case and do not end with punctuation, since they
  are usually wrapped: `open config: ...`, not `Failed to open config.`
- Use sentinel errors or custom error types only when callers need
  `errors.Is` or `errors.As`.
- Use early returns to reduce nesting.
- Do not panic for normal error handling.

## Context and Concurrency

- Accept `context.Context` as the first parameter for I/O, external calls,
  blocking work, and long-running operations.
- Do not store contexts in structs except for rare lifecycle-management cases.
- Use contexts for cancellation and timeouts.
- Manage goroutine lifecycles explicitly; avoid leaks.
- Use channels for communication and `sync` primitives when they are simpler
  or more appropriate.
- Avoid shared mutable state when possible.

## Testing, Makefile, and Services

These are specified once, in their own files:

- `tests.md`: the `make test` contract and how Go test suites are written.
- `../makefile/makefile.md`: the required Makefile targets (`make`, `test`,
  `build`, `clean`, `install`) and their contracts.
- `../services/`: when the program runs as a daemon or service, it ships
  `install` and `uninstall` subcommands and `make install` delegates to them.

## Security

- Validate all external input.
- Use `crypto/rand` for cryptographic randomness.
- Avoid command injection, SQL injection, path traversal, and unsafe handling
  of user-controlled data.
- Use prepared statements or equvilent to avoid including untrusted data in SQL or other database queries.
- Use least privilege.
- Do not log sensitive information.

## Performance

- Prefer clear code first; profile before optimising.
- Be mindful of allocations in hot paths.
- Use appropriate data structures.
- Prefer values over pointers where it improves clarity or performance.
- Use `sync.Pool` only when profiling or clear allocation patterns justify it.

## Style

- Use `gofmt` and `go vet`.
- Follow existing project conventions.
- Use `camelCase` for unexported names and `PascalCase` for exported names.
- Use descriptive names; avoid abbreviations except common ones such as `id`,
  `url`, and `http`.
- Keep initialisms in one case: `userID`, `baseURL`, `HTTPClient`, never
  `userId` or `HttpClient`.
- Use `New` for constructors.
- Keep functions focused.
- Prefer composition.
- Use pointer receivers for mutating methods.
- Keep receiver names short and consistent.
- Limit parameter lists; use structs for complex parameter sets.
- Document exported functions and types using godoc conventions.
- Do not manually edit generated files; update the source and regenerate.
