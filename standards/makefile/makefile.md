# Makefile Standards

This is how the author lays out Makefiles. It is a house convention, not an
industry standard; it is published because it works.

## Overview

Every project that builds on Linux and/or macOS has a `Makefile` at the
repository root exposing the targets below. The Makefile is the interface: the
same five commands mean the same thing in every project, to a developer, to a
CI/CD pipeline, and to an AI editor. How a target is implemented is the
project's business (Make itself, a script it calls, or both); the target names
and their contracts do not vary.

| Target | Contract |
| --- | --- |
| `make` | Runs `test`, then builds the binaries only if the tests pass. |
| `make test` | Runs the whole suite. Exits non-zero on any failure. The one gate, for developers and CI alike. |
| `make build` | Builds the binaries. Runs no tests. |
| `make clean` | Removes build output and test artifacts and drops the test cache. |
| `make install` | Installs the built binaries. Optional, but the user must be consulted before it is added or omitted. |

## `make` (the default target)

The default target is test-then-build: it depends on `test` and then on
`build`, in that order, so the binaries are produced only when the whole suite
has passed. A plain `make` either yields binaries that passed the tests or
exits non-zero before the build step runs.

Do not make `build` the default. A default that skips the tests turns `make`
into "compile it and hope"; the ordering exists so that the shortest command
is the safe one.

## `make test`

Runs the full regression suite and exits non-zero if any test fails. It is the
single gate: what a developer runs before declaring work done, and what the
CI/CD pipeline runs to decide whether anything proceeds. Any failure stops the
pipeline. There is no fast variant, and no section of the suite may be opted
out.

The full contract (what "whole suite" covers, no working-tree modification,
summary output, missing prerequisites, delegation to a root `test.sh`) is in
`../go/tests.md`. Projects in other languages meet the same contract with
their own tooling.

## `make build`

Compiles the binaries and nothing else: no tests, no install. It serves the
inner loop (edit, rebuild, run) and the `test` target when the suite needs the
binaries. Build flags, ldflags, and the build stamp are described in
`templates/go/app/README.md`.

## `make clean`

Returns the tree to the state where the next `make` starts from nothing:
deletes build output, test artifacts, and generated files, and clears the
language's test cache (for Go, `go clean -testcache`; see `../go/tests.md`).
`make clean; make` is a fresh build and `make clean; make test` is a fresh
run. It never touches tracked source files.

## `make install`

Installs the binaries produced by `build`, so it depends on `build` and never
installs a stale binary. The destination is decided when `make install` runs,
by who is running it, not by a per-project setting:

| Running as | Destination |
| --- | --- |
| root | `/usr/local/bin` (or the OS equivalent) |
| Not root, and `~/bin` exists | `~/bin` |
| Not root, and no `~/bin` | Error. Say that a system install needs `sudo make install`, and that a personal install needs `~/bin` to exist. Do not create it. |

A program that needs root to run (a system service, anything that binds a
privileged port or writes system paths) installs only as root. Run as a
normal user, it fails with an error that says to use `sudo`; it never falls
back to `~/bin`.

For a program that runs as a service, `make install` delegates to the binary's
own `install` subcommand rather than copying the binary itself, so the
destination and privilege rules exist in one place. See `../services/`.

The target is optional, but its absence is a decision the user makes, never
an oversight:

- If the Makefile has no `install` target, ask the user whether to add one
  and, if the program is not obviously one or the other, whether it can run
  without root. Do not silently leave the target out.
- The user may authorize omitting it (a library, a service deployed by other
  means, a tool run from the build tree). Record that decision in the project
  instructions file (`CLAUDE.md` or `AGENTS.md`) so the question is not asked
  again.

## Additional targets

Projects may add targets (`fmt`, `lint`, `build-all`, `test-webui`, and so
on). They never alter the contracts above: nothing else becomes the default,
and no other target is presented as the test gate.
