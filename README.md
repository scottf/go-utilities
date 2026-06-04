# go-utilities

Go ports of a couple of the Java utilities.

## sona — Sonatype checker

CLI port of the Java `SonatypeChecker`. For each configured project it fetches
the Maven metadata from the release repo (`repo1.maven.org`) and the snapshot
repo (`central.sonatype.com`) and reports when each artifact was last updated.

Projects are read from `~/.sonatype-checker.json` (a JSON array of
`CheckerParams`; the same format used by the Java side).

```
sona              # show the project menu, pick one (number, name, or 'a' for all)
sona 9            # run the 9th project in the list
sona FLINK        # run the project named FLINK (case-insensitive)
sona all          # run every project
sona -f other.json FLINK
sona -debug FLINK # print each URL before checking
```

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
```

Downloads the log for a GitHub Actions job (identified by its `github.com` job
URL) and prints the failing-test sections. The token is read from `$GH_TOKEN`
or `$GITHUB_TOKEN` (or pass `-token`).

```
export GH_TOKEN=ghp_...
ghhelp findJobFailures "https://github.com/nats-io/nats.java/actions/runs/25391554069/job/74467105487?pr=1564"
```

## Build

```
go build -o bin/sona   ./cmd/sona
go build -o bin/ghhelp ./cmd/ghhelp
go test ./...
```
