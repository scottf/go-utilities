# Changelog

## 0.1.0
- Initial release of two stdlib-only Go CLIs in one module
- `sona` — interactive Sonatype checker: fetches release metadata from
  `repo1.maven.org` and snapshot metadata from `central.sonatype.com`, reports
  when each artifact was last updated. Config in `~/.sonatype-checker.json`
  (override with `-f`), with `jdk_qualifiers` expansion (`null` = bare artifact)
- `ghhelp findJobFailures <jobUrl>` — downloads a GitHub Actions job log and
  prints the failing-test sections; token read from `$GH_TOKEN`/`$GITHUB_TOKEN`
  (or `-token`), `Authorization` stripped on cross-host redirect
- Ported from the Java originals in `java-utilities` (no hardcoded token carried over)
