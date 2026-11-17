/******************************************************************************
 * Copyright (c) [YEAR] [COMPANY]                                             *
 * Please see LICENSE file for details.                                       *
 ******************************************************************************/

// Package app holds the application's identity: name, short name, tagline,
// copyright and version. It is the only place any of them is defined.
//
// Values are unexported and read through accessors, so there is exactly one way
// to read each. The version is a const so the linker cannot override it; only
// build metadata is injected at link time.
package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	name      = "ExampleApp"
	shortName = "Example"
	tagLine   = "Does the thing"
	copyright = "Copyright (c) [YEAR] [COMPANY]"

	// version is bare semver. Build tooling reads this line; keep it a single
	// one-line assignment.
	version = "0.0.1"
)

// Build metadata, injected via ldflags:
//
//	-X <module>/app.gitCommit=<sha8>
//	-X <module>/app.buildTime=<rfc3339>
//	-X <module>/app.goVersion=<go version>
//	-X <module>/app.buildNumber=<utc yyyymmddhhmmss>
//
// Keep these as plain package-level vars: -X cannot set struct fields and
// fails silently if asked to.
var (
	gitCommit   string
	buildTime   string
	goVersion   string
	buildNumber string
)

// Name returns the product name, for identifying the product: titles, version
// lines, report headers, User-Agent strings.
func Name() string { return name }

// ShortName returns the informal product name, for running prose: log lines,
// alerts, messages and prompts ("Example is shutting down").
func ShortName() string { return shortName }

// BinaryName returns the name this program was started as: the last element
// of os.Args[0], without its directory and without a Windows ".exe" suffix.
// Use it wherever text tells someone what to type, so the instruction stays
// right when the binary is renamed or run through a link. It falls back to
// the lower-cased product name when the operating system gives no name.
func BinaryName() string {
	if len(os.Args) > 0 {
		b := filepath.Base(os.Args[0])
		if runtime.GOOS == "windows" && strings.EqualFold(filepath.Ext(b), ".exe") {
			b = b[:len(b)-len(".exe")]
		}
		if b != "" && b != "." && b != string(filepath.Separator) {
			return b
		}
	}
	return strings.ToLower(name)
}

// TagLine returns the one-line product description.
func TagLine() string { return tagLine }

// Copyright returns the copyright notice.
func Copyright() string { return copyright }

// Version returns the display form, "0.0.1+1a2b3c4d [20260902155301]", omitting
// whichever parts are not stamped.
//
// The commit is joined with "+" (SemVer build metadata, ignored in comparisons)
// so release and commit stay one token when the string is copied.
func Version() string {
	v := version
	if gitCommit != "" {
		v += "+" + gitCommit
	}
	if buildNumber != "" {
		v += " [" + buildNumber + "]"
	}
	return v
}

// SemVer returns the bare release number, for protocol handshakes and version
// comparisons. Use Version everywhere else.
func SemVer() string { return version }

// Build returns the build number (UTC link time as yyyymmddhhmmss), or "" if
// unstamped. Unlike the commit hash, it orders builds.
func Build() string { return buildNumber }

// BuildDate returns the UTC link day as YYYYMMDD, or 0 if unstamped. Use it only
// where an int is required; the full stamp does not fit in 32 bits.
func BuildDate() int {
	if len(buildNumber) < 8 {
		return 0
	}
	n, err := strconv.Atoi(buildNumber[:8])
	if err != nil {
		return 0
	}
	return n
}

// BuildInfo returns the build timestamp and Go toolchain version. The toolchain
// falls back to runtime.Version(); the timestamp stays empty if unstamped.
func BuildInfo() (string, string) {
	if goVersion == "" {
		return buildTime, runtime.Version()
	}
	return buildTime, goVersion
}
