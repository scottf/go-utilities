# go-utilities

Go ports of a couple of the Java utilities.

## sona — Sonatype checker

CLI port of the Java `SonatypeChecker`. For each configured project it fetches
the Maven metadata from the release repo (`repo1.maven.org`) and the snapshot
repo (`central.sonatype.com`) and reports when each artifact was last updated.

Projects are read from `~/.sonatype-checker.json` (a JSON array of
`CheckerParams`; the same format used by the Java side).

```
sona                       # print help, then the config location and project list
sona FLINK                 # check the project named FLINK (case-insensitive)
sona -c other.json FLINK   # use a different config file
sona -d FLINK              # debug: print each URL before checking
```

Command shape: `sona [-d] [-c path-to-config] project-name`. The `-d` and `-c`
flags may appear in any order; the project name is the lone non-flag argument.

Config entry shape:

```json
{
  "project": "FLINK",
  "group_id": "io.synadia",
  "component": "flink-connector-nats",
  "jdk_qualifiers": [null],
  "release_versions": ["3.0.3", "2.3.1", "2.3.2"],
  "snapshot_versions": ["3.0.4", "2.3.2", "2.3.3"]
}
```

A `null` (or empty) entry in `jdk_qualifiers` means the bare artifact with no
`-jdkNN` suffix; e.g. `[null, "17", "21", "25"]` checks `component`,
`component-jdk17`, `component-jdk21`, and `component-jdk25`.

## ghhelp — GitHub Actions helper

```
ghhelp findJobFailures <jobUrl>
ghhelp fjf <jobUrl>              # shortcut for findJobFailures
fjf <jobUrl>                    # standalone command, same as 'ghhelp fjf'
```

Downloads the log for a GitHub Actions job (identified by its `github.com` job
URL) and prints the failing tests — one stack trace per **distinct test**.
Different tests are never combined; only repeats of the *same* test (e.g. the
same test failing across several matrix builds in one log) are collapsed into a
single block with an `(xN)` count, since those repeats are identical. Pass `-v`
to print every failure block, including duplicates.

The token is read from `$GH_TOKEN` or `$GITHUB_TOKEN` (or pass `-token`).

```
export GH_TOKEN=ghp_...
fjf "https://github.com/nats-io/nats.java/actions/runs/25391554069/job/74467105487?pr=1564"
fjf -v "https://github.com/..."   # full stack traces instead of the summary
```

Quote the URL. It works unquoted in most shells, but the `?` and any `&` in the
query string are shell metacharacters, so quoting keeps it safe across shells
and for URLs with multiple query parameters.

## fjf — findJobFailures shortcut

`fjf` is a standalone command that does exactly what `ghhelp fjf` (i.e.
`ghhelp findJobFailures`) does, so you can skip the subcommand:

```
fjf <jobUrl>          # summary: one stack trace per distinct failing test
fjf -v <jobUrl>       # full output: every failure block, including duplicates
fjf -token tok <url>  # pass a token explicitly instead of via the environment
```

Command shape: `fjf [-v] [-token tok] <jobUrl>`. Output, token resolution
(`$GH_TOKEN` / `$GITHUB_TOKEN` / `-token`), and URL quoting are identical to
[`ghhelp`](#ghhelp--github-actions-helper) above.

## Build

```
go build -o bin/sona   ./cmd/sona
go build -o bin/ghhelp ./cmd/ghhelp
go build -o bin/fjf    ./cmd/fjf
go test ./...
```

## Release

Releases are built by the GitHub Actions workflow in
`.github/workflows/release.yml`, which fires when a GitHub release is published.
It cross-compiles `sona`, `ghhelp`, and `fjf` for linux/amd64, linux/arm64,
darwin/amd64, darwin/arm64, and windows/amd64, attaches the binaries to the
release, and uses `CHANGELOG.md` as the release notes.

To cut a release:

1. Add a section for the new version at the top of `CHANGELOG.md` and push it to
   `main`. The workflow reads the changelog from the tagged commit, so this has
   to be in place before the tag.
2. Create the release (this creates the tag from `main` and triggers the build):

   ```
   gh release create v0.1.0 --title "v0.1.0" --notes-file CHANGELOG.md
   ```

3. Watch it and confirm the assets:

   ```
   gh run watch
   gh release view v0.1.0
   ```

Bump the version per release (`v0.2.0`, `v0.3.0`, …) with a matching
`CHANGELOG.md` section. The binaries don't embed a version string, so the git
tag and changelog are the source of truth.
