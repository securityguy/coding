# Service and Daemon Standards — Windows

## Overview

A program that runs as a service (anything the operating system starts and
keeps running) ships with its own `install` and `uninstall` subcommands. The
user goes from a built binary to a running, enabled service with one command,
and back again with one command. A README that walks the user through
registering the service by hand is not a substitute; the program knows what it
needs and does it.

Adding the two subcommands is the default. Their absence is a decision the
user makes, never an oversight:

- When writing a new service, ask the user the questions in
  "What to ask" below. Do not ask whether to have `install` and `uninstall`
  at all; ask how they should behave.
- The user may authorize omitting them (a service deployed only by a
  package or container, a component started by a parent process). Record
  that decision in the project instructions file (`CLAUDE.md` or
  `AGENTS.md`) so it is not asked again.

A program that also ships on Linux or macOS follows `linux.md` or `macos.md`
for those platforms.

## Subcommands

| Command | Contract |
| --- | --- |
| `progname install` | Copies the running binary to the install location, registers the service with the Service Control Manager, sets it to start at boot, and starts it. Running it over an existing install is an upgrade: replace the binary, update the registration, restart. |
| `progname uninstall` | Stops the service, removes it from the Service Control Manager, and removes the installed binary. Leaves configuration and data in place and prints where they are. Not installed: say so, exit 0. |

`install` copies *itself* (the executable that is running) so that running
`progname install` from an elevated prompt in the build tree is the whole
procedure. Both subcommands print each step they take and exit non-zero on the
first failure, naming the file or command that failed.

If the service cannot start until it is configured, `install` still installs
and enables it, then says so and prints the one command that starts it once
configuration is in place.

## Scope

There is no user scope on Windows; services are registered system-wide and
require an elevated (Administrator) prompt. Run without elevation: exit with an
error that says to rerun from an elevated prompt. There is no `--user` flag.

## What the registration contains

The registration that `install` creates is the reference for how the program
runs. It includes at least:

- The full path to the installed binary and the arguments it needs.
- The account it runs as. A service that does not need full privileges runs
  as a dedicated unprivileged account, which `install` creates if missing.
- Restart on failure.
- The working directory and any environment the program requires.

## What to ask

Before writing a service, settle these with the user and record the answers
in the project instructions file (`CLAUDE.md` or `AGENTS.md`):

- Does the program need full privileges to run, or can it run unprivileged?
  This decides what account the service runs as.
- Which platforms does it target? Write the definition for each one it
  ships on and no others.
- Is there a reason to omit `install` and `uninstall` entirely?
