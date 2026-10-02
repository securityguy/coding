# Go Testing Standards

## Overview

Every Go project must have a `make test` target that executes the full regression
suite and summarises the results. It is the single entry point for testing: what a
developer runs locally, and what a pipeline runs to decide whether a build may proceed.

if unable to comply with the makefile requirements due to p,atform limitations,
advise the user and request clarification.

There is deliberately one target, not a fast `test` and a full `check`. Two
targets invite treating the fast one as the gate. The inner loop while
iterating is `go test ./some/package` run directly; `make test` is what "the
tests pass" means.

`make test` is the interface. How it is implemented is the project's
business: a Makefile that does the work itself, or one that calls a `test.sh`
at the project root, or both. Where the suite is long or needs orchestration a
script is usually the better home for it. Make is a poor language for argument
parsing, colour handling, ephemeral services and cleanup traps. A project
keeping a `test.sh` must still expose it through `make test`, so there is one
command to remember and one thing for CI to call.

The other Makefile targets and their contracts are in
`../makefile/makefile.md`.

## The `make test` Contract

- Must run the **whole** suite. If the project has unit tests, integration
  tests, a frontend and a linter, `make test` covers all of them. A gate that
  leaves a section out means "I ran the tests" says different things to
  different people.
- Must exit `0` on full success, non-zero on any failure.
- Must print a summary showing total, passed, failed, and skipped counts.
- **Must not modify the working tree.** Formatting is *verified*, not
  applied: `golangci-lint fmt --diff` rather than `golangci-lint fmt`. Keep
  the rewriting form as a separate `make fmt`. A gate that edits the code it
  is judging cannot be trusted in CI.
- Must need no running service, no manual setup, and no interactive input.
- Must be self-contained for project-local build steps: it builds the
  binaries it needs, but must not install system packages or global tools
  without explicit user approval.
- Must clean up test artifacts on completion (offer a flag to preserve them
  for debugging).
- If the suite cannot run in full because of a missing external prerequisite,
  it must exit non-zero and name the prerequisite.

Where part of the suite genuinely cannot meet those terms (a browser suite
needing a running instance, say) give it its own target (`make test-webui`)
and document when to run it. Do not fold it into `test` and do not leave it
undiscoverable.

If an assistant cannot run `make test`, it must report the exact command
attempted, the reason it failed, and what remains unverified.

**All tests must execute.** A passing exit code is only meaningful if every
test section ran. If any section is silently skipped, the suite result is
unreliable. When reporting "all tests passed", confirm that no section was
skipped. Missing external prerequisites (e.g. a required binary) must cause a
non-zero exit, not a silent skip.

## `make clean` Drops the Test Cache

`go test` replays a package's last passing result as `(cached)` when nothing
it can see has changed. That is normally what you want, but a test that reads
time, randomness, the network or files outside its own opens can be replayed
when it should have run. `make test` bypasses the cache with `-count=1`;
`make clean` must also run `go clean -testcache`, so a direct `go test ./...`
after a clean is a fresh run too:

```make
clean:
	rm -rf bin data
	go clean -testcache
```

## Shell Script Expectations

These apply to any script `make test` calls.

- Use portable shell where practical. Prefer `#!/usr/bin/env bash` when Bash
  features are used.
- Use explicit error handling. `set -u` is encouraged. Avoid relying on
  `set -e` when it would prevent collecting failures and printing the final
  summary.
- Scripts that start processes must use cleanup traps and bounded
  waits/timeouts. Do not leave services, sockets, temporary directories, or
  test data behind.
- Never suppress errors from `go test` or other sub-commands. If a command
  fails, propagate that failure to the final exit code.

## Foundation: `go test`

All Go projects use `go test` as the foundation for unit and integration
tests. The suite must invoke it with at minimum:

```sh
go test -race -count=1 ./...
```

- `-race`: enables the race detector. Mandatory for the default invocation;
  data races are real bugs.
- `-count=1`: disables test result caching so tests always run fresh.
- `./...`: tests all packages.

Convenience flags (e.g. `-f` for fast mode) may skip the race detector to
support quick local iteration. When used, the script must print a visible
warning that the run is incomplete. The default invocation, with no flags as
called by CI, must always include `-race`. Convenience flags are never
acceptable in a build chain.

For CI or when a parseable output format is needed, use `gotestsum` in place
of `go test`:

```sh
gotestsum --format testname --junitfile test-results.xml -- -race -count=1 ./...
```

`gotestsum` is a drop-in replacement with better output formatting and JUnit
XML support. It is recommended for projects integrated into a build chain.
Install via `go install gotest.tools/gotestsum@latest`.

## Test Structure

- Place tests in `_test.go` files in the same package as the code under test.
- Use descriptive test names.
- Write table-driven tests for scenarios with multiple inputs.
- Use the standard library for assertions. Add an assertion library such as
  `testify` only when the standard library is insufficient and the user has
  approved the dependency.
- Mock external dependencies at package boundaries; interfaces make this
  straightforward.
- Aim for meaningful coverage of critical paths; do not chase a coverage
  number.
- Do not write tests that only pass because a mock behaves the way you
  expect. Test real behaviour at integration boundaries.

## Integration and End-to-End Tests

For applications that expose an external interface (HTTP API, MCP server, CLI
tool), `go test` alone is insufficient. The suite also runs integration or
end-to-end tests after the unit tests pass.

- **MCP servers:** build the binary, call each tool through an MCP client with
  representative inputs, verify the expected output appears in the response,
  and report pass/fail per tool call.
- **HTTP APIs and services:** start the service against a test
  configuration, issue requests with `curl` or a purpose-built test client,
  and assert on response bodies and status codes.
- **CLI tools:** invoke the compiled binary with representative arguments and
  assert on stdout, stderr, and exit codes.

## Test Numbering Convention

For scripts with many tests, use a hierarchical numbering scheme to make
failures easy to locate:

```
0.x   - Setup and preconditions
1.x   - <Feature group 1>
2.x   - <Feature group 2>
...
N.x   - Error handling and edge cases
N+1.x - Cleanup and final verification
```

Use subsection numbers (e.g. `1.2.3`) when a feature group has sub-groupings.

## Summary Output Format

The final output of `make test` must include a summary block in this format:

```
============================================
   TEST SUMMARY
============================================

Total Tests: <n>
Passed:      <n>
Failed:      <n>
Skipped:     <n>

All tests passed!   ← or →   FAILURES DETECTED
```

`Total Tests` may be the count of test packages executed. Parsing individual
test case counts from `go test` output is not required; package-level
granularity is acceptable.

Use terminal colours where the environment supports them (green for pass, red
for fail). The script must detect colour support and disable escape codes when
stdout is not a terminal, when the `NO_COLOR` environment variable is set, or
when a `-n` flag is passed.

## Exit Code Contract

The exit code is the single interface between `make test` and any automated
system.

| Exit code | Meaning |
| --- | --- |
| `0` | All tests passed. Build may proceed. |
| Non-zero | One or more tests failed. Build must stop. |

**The suite must never exit `0` if any test failed.** This is what makes it
safe to use in a build chain without any special integration: every CI
system, Makefile, and shell pipeline treats a non-zero exit code as failure
and stops.

The human-readable summary exists for the developer. The exit code exists for
everything else. Both are always produced; they are not alternatives.

A correct script ends with something like:

```sh
if [ "$FAIL_COUNT" -gt 0 ]; then
    exit 1
fi
exit 0
```

## Build Chain Integration

When integrated into a CI/CD pipeline:

1. The pipeline calls `make test` with no arguments.
2. A non-zero exit code fails the build; no further configuration is
   required.
3. If using `gotestsum`, the JUnit XML file (`test-results.xml`) may be
   published as a test artifact for reporting.
4. Builds that fail tests do not proceed to packaging or deployment.
