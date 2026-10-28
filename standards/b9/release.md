# b9 Release Files

`b9` (github.com/tenebris-tech/b9builder) takes a project from source to a
signed, notarized, published GitHub release. It is driven by one file in the
project, `b9.yaml`. This document covers what a project must provide and what
goes in that file. Setting up a build host (certificates, YubiKey,
`~/.b9/config`) is covered by the b9builder README and is not repeated here.

## Where the release file lives

Every released project carries one file, `b9.yaml`, at the root of its
repository. It is committed and pushed with the source, so it describes the
product and never the machine: no absolute paths, no signing identity, no
notary profile, no work directory, no secrets. Those are host facts and come
from the environment or `~/.b9/config` on the build host.

The file is driven with `b9 verify`, `b9 build` and `b9 release`, each taking
the repository directory, the file path, or nothing when run from inside the
repository. Add `--dry-run` to print every command without executing any.

## What a project must provide

A project is ready for a `b9.yaml` when all of the following hold. Check them
before writing the file; `b9 verify` reports the ones it can detect.

- **a. A version constant in source.** The version is read out of a committed
  file by regex, never from git. Go projects keep it in `app/app.go` as
  `version = "x.y.z"` (see `go/writing.md`, Versioning and Copyright).
- **b. `make build PLATFORM=<goos> ARCH=<goarch>`.** Both variables default
  to the host. The target writes each binary to a predictable path that
  includes the OS and architecture, with `.exe` appended on Windows, and
  builds with `-trimpath` and `CGO_ENABLED=0`. b9 refuses a binary built
  without `-trimpath`. The Makefile block is below; the ldflags it uses are in
  `templates/go/app/README.md`.
- **c. `make clean`.** Run once before the first target when
  `build.clean_once` is set.
- **d. A `LICENSE` file.** Declare it in `package.extra_files` so it is copied
  into every archive as `LICENSE.txt`.
- **e. `license.tmpl`**, a `go-licenses` template, when the open source
  licence report is enabled. Copy it from an existing project unchanged.
- **f. A clean working tree and a free tag.** b9 refuses a dirty tree
  (`--allow-dirty` overrides) and a tag that already exists locally or on
  `origin`.
- **g. The GitHub repository exists** and `gh` is authenticated on the host.

### Makefile block

```makefile
BUILD_DIR=dist
UNAME_S=$(shell uname -s)
UNAME_M=$(shell uname -m)

# PLATFORM maps to GOOS and ARCH to GOARCH. Both default to the host, so a
# plain `make build` is unchanged; b9 overrides them per target.
ifeq ($(UNAME_S),Darwin)
PLATFORM?=darwin
else ifeq ($(UNAME_S),Linux)
PLATFORM?=linux
else
PLATFORM?=$(UNAME_S)
endif

ifeq ($(UNAME_M),x86_64)
ARCH?=amd64
else ifeq ($(UNAME_M),aarch64)
ARCH?=arm64
else ifeq ($(UNAME_M),arm64)
ARCH?=arm64
else
ARCH?=$(UNAME_M)
endif

EXE=$(if $(filter windows,$(PLATFORM)),.exe)
BUILD_ENV=GOOS=$(PLATFORM) GOARCH=$(ARCH)

build:
	@mkdir -p $(BUILD_DIR)
	$(BUILD_ENV) CGO_ENABLED=0 go build -trimpath $(LDFLAGS) \
		-o $(BUILD_DIR)/foo-$(PLATFORM)-$(ARCH)$(EXE) .
```

The output path chosen here is the `artifacts[].from` path in the release
file: `'<BUILD_DIR>/foo-{os}-{arch}{exe}'`.

## The file

```yaml
name:   Foo                          # product name
module: github.com/tenebris-tech/foo
github: tenebris-tech/foo            # owner/repo

version:
  file:    app/app.go
  pattern: 'version\s*=\s*"([^"]+)"'
  format:  '{1}'

go_min:      "1.27"
govulncheck: true

licenses:
  enabled:  true
  template: license.tmpl
  ignore:   github.com/tenebris-tech/foo
  out:      OPEN-SOURCE-LICENSES.txt

build:
  clean_once: true
  command: 'make build PLATFORM={os} ARCH={arch}'
  targets:
    - windows/amd64
    - windows/arm64
    - linux/amd64
    - linux/arm64
    - darwin/arm64

artifacts:
  - name: foo
    from: 'dist/foo-{os}-{arch}{exe}'
    macos_bundle_id: com.tenebris.foo
    windows_sign: true

package:
  extra_files:
    - {from: LICENSE, as: LICENSE.txt}
    - {from: '{staging}/OPEN-SOURCE-LICENSES.txt', as: OPEN-SOURCE-LICENSES.txt}
  per_platform:
    enabled: true
    # CONTRACT: install scripts fetch these exact names from /releases/latest/download/.
    asset:         'foo-{os}-{arch}.tar.gz'
    windows_asset: 'foo-{os}-{arch}.zip'
  bundle:
    enabled: false

release:
  tag:   '{version}'
  title: 'v{tag}'
  notes: 'Windows, Linux and macOS binaries (Windows binaries are signed; macOS binaries are signed and notarized)'
```

Unknown keys are rejected: a misspelled `windows_sign` would otherwise ship an
unsigned binary.

### Fields

Required fields are marked. A field that is required only in some
configurations says so.

| Field | Meaning |
| --- | --- |
| `name` | Product name. Required. |
| `module` | Go module path. Documentation only; b9 does not read it. |
| `github` | `owner/repo` of the GitHub repository that receives the release. Required. |
| `repo` | Source tree relative to the file. Only for a conf kept outside its repository; an absolute path in an in-repo `b9.yaml` is rejected. |
| `version.file` | Committed file holding the version, relative to the repository. Required. |
| `version.pattern` | Regex with one capture group for the version. Required. |
| `version.format` | Version string built from `{1}` and any `version.extra` keys. Required. |
| `version.extra` | Extra named regexes over the version file, each with one capture group. |
| `go_min` | Minimum Go version, compared numerically. Required. |
| `govulncheck` | Run `govulncheck ./...` once before building. Findings are fatal. |
| `licenses.enabled` | Generate an open source licence report with `go-licenses` into staging. |
| `licenses.template` | The `go-licenses` template, relative to the repository. Required when enabled. |
| `licenses.ignore` | Module path to leave out of the report, normally the project's own. |
| `licenses.out` | Report filename in staging. Required when enabled. |
| `build.command` | Command run once per target, in the repository. Required. |
| `build.clean_once` | Run `make clean` once before the first target, never between targets. |
| `build.targets` | `goos/goarch` pairs, no duplicates. Required, non-empty. This list is the complete set of release platforms. |
| `artifacts[].name` | Bare binary name as the user runs it. Required; unique. |
| `artifacts[].from` | Where the build wrote this target's binary, relative to the repository. Deleted before each target builds. Required. |
| `artifacts[].macos_bundle_id` | Bundle identifier used as the codesign identifier. Required when any `darwin` target exists. |
| `artifacts[].windows_sign` | Sign this binary's Windows builds with jsign and the YubiKey. Set it on every Windows binary that ships. |
| `artifacts[].bundle_dir` | Subdirectory inside the all-platform bundle. Empty means the bundle root. |
| `signing.macos.identity`, `.notary_profile` | Accepted by the schema as checked-in defaults for the codesign path. Host facts; never set them in our files. |
| `signing.macos.smartcard_slot` | PIV slot holding the Developer ID key, for rcodesign on a smartcard. The environment wins. |
| `signing.windows.*` | Checked-in defaults for jsign (`jsign`, `slot`, `alias`, `tsaurl`, `pin_env`). The environment wins. Normally omitted. |
| `signing.minisign.enabled` | Publish a minisign-signed `checksums.txt` beside the archives. |
| `signing.minisign.public_key` | The base64 line of the minisign public key the product embeds. Required when enabled. |
| `signing.minisign.checksums` | Checksum list filename, a bare literal. Default `checksums.txt`. |
| `signing.minisign.trusted_comment` | Signed comment template. Default `{name} v{version}`. |
| `package.extra_files[]` | `{from, as}` pairs copied into every archive. `from` is relative to the repository unless it uses `{staging}`. |
| `package.per_platform.enabled` | One archive per target. This or `bundle` must be enabled. |
| `package.per_platform.asset` | Archive name for non-Windows targets; must contain `{os}` and `{arch}`. Required when enabled. |
| `package.per_platform.windows_asset` | Archive name for Windows targets, a `.zip`. Required when enabled and a Windows target exists. |
| `package.bundle.enabled` | One archive holding every platform. |
| `package.bundle.asset` | Bundle archive name. Required when enabled. |
| `package.bundle.root_dir` | Top-level directory inside the bundle. Required when enabled. |
| `release.tag` | Git tag template. Required. |
| `release.title` | Release title template. Required. |
| `release.notes` | Release notes text. |

### Placeholders

`{key}` is substituted per field. An unknown placeholder is left in place
rather than blanked, so a typo shows up in the asset name instead of leaving a
hole.

| Field | Available |
| --- | --- |
| `build.command`, `artifacts[].from` | `{os}` `{arch}` `{exe}` |
| `package.per_platform.asset`, `.windows_asset` | `{os}` `{arch}` `{exe}` `{version}` `{tag}` |
| `package.bundle.asset` | `{version}` `{tag}` |
| `package.extra_files[].from` | `{staging}` |
| `version.format` | `{1}` and each `version.extra` key |
| `release.tag` | `{version}` |
| `release.title` | `{version}` `{tag}` |
| `signing.minisign.trusted_comment` | `{name}` `{version}` |

`{exe}` is `.exe` on Windows and empty elsewhere.

## Choosing the packaging

- **Per-platform archives** are the default: a `.tar.gz` per Unix target and
  a `.zip` per Windows target, each holding the binaries under their bare
  `name` plus the extra files. A user downloads one archive and runs
  `./foo`. Windows archives are `.zip` because that is what Windows opens
  without a tool.
- **An all-platform bundle** is for a product whose installer needs every
  platform's binaries at once (a server that deploys agents, for example).
  Binaries sit under `<root_dir>/<bundle_dir>/<name>-<os>-<arch>{exe}`.
- Both can be enabled in the same file.

**Asset names are contracts.** Install and update scripts already running on
other people's machines fetch these filenames, and a deployed script requests
the old name forever. Declare them literally, keep them version-free so
`/releases/latest/download/<asset>` resolves, and never rename one after a
release has shipped. Pinning to a version works through
`/releases/download/<tag>/<asset>`. Mark the lines with a `CONTRACT` comment.

## Signing and notarization

b9 never modifies the files the build wrote. Each artifact is copied into a
staging directory under its bare name (`foo`, or `foo.exe` on Windows), and
signing, notarization and packaging work on those copies.

- **macOS.** Every artifact built for a `darwin` target is signed with the
  hardened runtime, a secure timestamp and `macos_bundle_id` as the
  identifier, then notarized in one submission per release. On a macOS build
  host it is also checked with Gatekeeper's own assessment. Nothing modifies
  a macOS binary after notarization. All of this follows from a `darwin`
  target and a bundle ID; there is nothing else to enable. Use
  `com.tenebris.<name>` for the bundle ID.
- **Windows.** Set `windows_sign: true` on every artifact that ships for a
  `windows` target. The binary is signed with jsign through the YubiKey,
  timestamped, and verified with `osslsigncode`. Windows has no notarization
  step; a timestamped Authenticode signature is the complete requirement.
- **Linux.** Not signed. A product with its own updater signs the checksum
  list instead (next item).
- **minisign.** Enable `signing.minisign` only when the product has an
  updater that verifies releases; the product embeds the public key and
  refuses a release whose `checksums.txt` is not signed by it. b9 verifies
  its own signature against `public_key` before publishing, so a wrong host
  key fails the run. A tool installed by hand or by a package manager does
  not need it.

## Rules

- A `b9.yaml` describes a product, never a machine. Anything host-specific is
  an environment variable or a line in `~/.b9/config`.
- No secrets in the file: no PIN, no password, no key path.
- The version comes from source, never from `git describe`.
- `build.targets` is the complete list of release platforms. Build, vet and
  test only those; do not add or "fix" other platforms unless asked. When
  the list changes, update the project's `CLAUDE.md` and `test.sh` to match.
- `make` is the build interface. `build.command` can call `go build`
  directly while a project has no Makefile, but `make build PLATFORM= ARCH=`
  is the target state.
- Adding a project to b9 means writing a `b9.yaml`, not changing b9. If a
  project cannot be described, the schema is missing something; extend the
  schema.

## Running a release

```sh
b9 verify            # preflight only: tools, credentials, version, tag, tree
b9 build             # build, sign, notarize, package; publishes nothing
b9 release           # all of the above, then tag, push, GitHub release
b9 release --tag rc1 # pre-release: "-rc1" appended to the tag, marked pre-release
b9 release --dry-run # print every command, run none
```

Run these from inside the repository, or name it: `b9 release ~/source/foo`.
A `build` leaves its assets in the `assets/` subdirectory of the run directory
under `~/.b9/tmp` and prints the path; a successful `release` uploads them and
removes the directory; a failed run of either kind keeps it for inspection.
Nothing is pushed unless every earlier stage passed.
