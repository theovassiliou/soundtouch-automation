package main

import (
	"fmt"
	"strings"
)

// set this via ldflags (see https://stackoverflow.com/q/11354518)
const pVersion = ".3"

// version is the current version number as tagged via git tag 1.0.0 -m 'A message'
var (
	version = "1.1" + pVersion + "-src"
	commit  string
	branch  string
	build   string
)

// FormatFullVersion formats for a cmdName the version number based on version, branch and commit
// and adds an optional, i.e. non-empty build id
// A resulting full version looks like:
// myApp 1.0 (git: main b2fecc) (build: 2023-01-02T14:22:23Z)
func FormatFullVersion(cmdName, version, branch, commit, build string) string {
	var parts = []string{cmdName}

	if version != "" {
		parts = append(parts, version)
	} else {
		parts = append(parts, "unknown")
	}

	if branch != "" || commit != "" {
		if branch == "" {
			branch = "unknown"
		}
		if commit == "" {
			commit = "unknown"
		}
		git := fmt.Sprintf("(git: %s %s)", branch, commit)
		parts = append(parts, git)
	}

	if build != "" {
		build := fmt.Sprintf("(build: %s)", build)
		parts = append(parts, build)
	}
	return strings.Join(parts, " ")
}
