# go-utilities — session handoff

Context note for picking this up in a fresh Claude session rooted in
`/mnt/c/dev/go-utilities`. These tools were ported from the Java originals in
`/mnt/c/dev/java-utilities` during a session that was rooted in that repo.

## What's here

Two stdlib-only CLIs in one module (`module go-utilities`, Go 1.24, no external deps):

```
go.mod
cmd/sona/main.go                 "sona" CLI — interactive Sonatype checker
cmd/ghhelp/main.go               "ghhelp" CLI — GitHub Actions helper
internal/sonatype/sonatype.go    config load, maven-metadata.xml parse, check logic
internal/sonatype/sonatype_test.go
internal/ghhelp/ghhelp.go        FindJobFailures
internal/ghhelp/ghhelp_test.go
README.md
```

Build / verify:

```
go build ./...
go vet ./...
go test ./...        # both internal packages have passing tests
```

## sona

Port of Java `scottf.sc.SonatypeChecker` + `CheckerParams`. For each project it
fetches release metadata from `repo1.maven.org` and snapshot metadata from
`central.sonatype.com`, and prints when each artifact was last updated.

- Config: `~/.sonatype-checker.json` (copied from
  `java-utilities/src/main/java/scottf/main/SonatypeCheckerParams.json`).
  Resolved via `os.UserHomeDir()`, so it works from WSL or Windows. Override
  with `-f path`.
- JSON uses snake_case keys (`group_id`, `jdk_qualifiers`, `release_versions`,
  `snapshot_versions`) matching the Java `@JsonProperty` names.
- A `null`/empty entry in `jdk_qualifiers` = the bare artifact (no `-jdkNN`
  suffix). `[null,"17","21","25"]` checks component, component-jdk17, -jdk21, -jdk25.
- Timestamps are formatted as RFC3339 UTC (e.g. `2026-06-04T14:25:43Z`) to match
  the Java `ZonedDateTime` output.

Run:

```
sona            # menu: pick by number, name, or 'a' for all
sona FLINK      # by name (case-insensitive)
sona 9          # by list position
sona all
sona -debug FLINK
```

Verified live this session: FLINK and AP (snapshot-only) runs produced correct
output including jdk-qualifier expansion.

## ghhelp

Port of `findJobFailures` from Java `scottf.GitHubApiRequest`.

```
ghhelp findJobFailures <jobUrl>
```

Downloads a GitHub Actions job log and prints the failing-test sections (same
`() FAILED` … until `() STARTED`/`tests completed,` windowing, with the leading
ISO-timestamp stripped). Handles both job-URL shapes (with/without `?pr=`).

Token: read from `$GH_TOKEN` or `$GITHUB_TOKEN` (or `-token`). `GH_TOKEN` was
added to `~/.bashrc` this session.

## Deliberate departures from the Java source

1. **No hardcoded token.** Java `GitHubApiRequest.main` had a live `ghp_...`
   token embedded; it was NOT copied. ghhelp reads it from the environment.
2. **Auth header on redirect.** GitHub's job-logs endpoint 302-redirects to
   signed blob storage; the Go client strips `Authorization` when the redirect
   leaves `api.github.com`, mirroring the manual handling in Java `executeRaw`.

## Open items / TODO

- [ ] `findJobFailures` not yet run end-to-end (needs a token + a real failed
      job URL). `parseJobURL` is unit-tested against both example URLs.
- [ ] SECURITY: the `ghp_...` token is committed in plaintext in
      `java-utilities/.../GitHubApiRequest.java` (and now in `~/.bashrc`). It
      should be rotated; consider scrubbing it from the Java source and reading
      from env there too.
- [ ] No `go.sum` / external deps yet — the interactive menu is plain stdin. If
      a richer TUI is wanted later, that'd add a dependency.
```
