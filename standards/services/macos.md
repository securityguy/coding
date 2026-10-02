# Service and Daemon Standards — macOS

## Overview

A program that runs as a daemon, service, or agent (anything the operating
system starts and keeps running) ships with its own `install` and `uninstall`
subcommands. The user goes from a built binary to a running, enabled service
with one command, and back again with one command. A README that walks the
user through writing a plist by hand is not a substitute; the program knows
what it needs and writes it.

Adding the two subcommands is the default. Their absence is a decision the
user makes, never an oversight:

- When writing a new service, ask the user the questions in
  "What to ask" below. Do not ask whether to have `install` and `uninstall`
  at all; ask how they should behave.
- The user may authorize omitting them (a service deployed only by a
  package or container, a component started by a parent process). Record
  that decision in the project instructions file (`CLAUDE.md` or
  `AGENTS.md`) so it is not asked again.

A program that also ships on Linux or Windows follows `linux.md` or
`windows.md` for those platforms.

## Subcommands

| Command | Contract |
| --- | --- |
| `progname install` | Copies the running binary to the install location, writes the launchd plist, loads it so it starts at boot, and starts it. Running it over an existing install is an upgrade: replace the binary, rewrite the plist, restart. |
| `progname uninstall` | Stops and unloads the service and removes the plist and the installed binary. Leaves configuration and data in place and prints where they are. Not installed: say so, exit 0. |

`install` copies *itself* (the executable that is running) so that
`sudo ./progname install` from the build tree is the whole procedure. Both
subcommands print each step they take and exit non-zero on the first failure,
naming the file or command that failed.

If the service cannot start until it is configured, `install` still installs
and loads it, then says so and prints the one command that starts it once
configuration is in place.

## System or user scope

launchd offers a system scope (LaunchDaemons), which needs root, and a user
scope (LaunchAgents), which does not. The default is system scope.

| Scope | Plist | Binary |
| --- | --- | --- |
| System (default) | `/Library/LaunchDaemons/<id>.plist`; run as `sudo progname install` | `/usr/local/bin` |
| User (`--user`) | `~/Library/LaunchAgents/<id>.plist` | `~/bin` |

System scope is the default because a user-scope service only runs while that
user has a session. On a headless machine there is none, and a service that
silently fails to come up after a reboot is worse than one that asked for
`sudo` up front. Requiring `sudo progname install` is reasonable.

- Run without root and without `--user`: exit with an error that says to rerun
  under `sudo`, or to pass `--user` if the program supports it.
- `--user` exists only when the program can do its job without privileges
  (no privileged ports, no system paths). If it cannot, there is no `--user`
  flag and the error says root is required.
- `uninstall` takes the same `--user` flag and removes what the matching
  `install` created.

## What the plist contains

The plist that `install` writes is the reference for how the program runs. It
includes at least:

- The full path to the installed binary and the arguments it needs.
- The account it runs as. A service that does not need root runs as a
  dedicated unprivileged account, which `install` creates if missing.
- Restart on failure.
- The working directory and any environment the program requires.

Keep it the plist itself, not a template for the user to fill in.

## Relationship to `make install`

`make install` (see `../makefile/makefile.md`) installs binaries; it does not
register services. For a service, `make install` delegates to the binary's own
`install` subcommand so the destination and privilege rules live in one place.
Running it as root installs and loads the system service; running it as a
normal user installs a user-scope service if the program has `--user`, and
otherwise fails with the same error the subcommand gives.

## What to ask

Before writing a service, settle these with the user and record the answers
in the project instructions file (`CLAUDE.md` or `AGENTS.md`):

- Does the program need root to run, or can it run unprivileged? This decides
  whether `--user` exists and what account the system service runs as.
- Which platforms does it target? Write the definition for each one it
  ships on and no others.
- Is there a reason to omit `install` and `uninstall` entirely?
