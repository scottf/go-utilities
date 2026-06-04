package sonatype

import "testing"

const releaseXML = `<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>io.nats</groupId>
  <artifactId>jnats-server-runner</artifactId>
  <versioning>
    <latest>3.1.0</latest>
    <release>3.1.0</release>
    <versions>
      <version>3.0.2</version>
      <version>3.1.0</version>
    </versions>
    <lastUpdated>20260127204946</lastUpdated>
  </versioning>
</metadata>`

const snapshotXML = `<?xml version="1.0" encoding="UTF-8"?>
<metadata modelVersion="1.1.0">
  <groupId>io.synadia</groupId>
  <artifactId>flink-connector-nats</artifactId>
  <version>3.0.4-SNAPSHOT</version>
  <versioning>
    <snapshot>
      <timestamp>20260602.193358</timestamp>
      <buildNumber>4</buildNumber>
    </snapshot>
    <lastUpdated>20260602193358</lastUpdated>
    <snapshotVersions>
      <snapshotVersion>
        <classifier>sources</classifier>
        <extension>jar</extension>
        <value>3.0.4-20260602.193358-4</value>
        <updated>20260602193358</updated>
      </snapshotVersion>
    </snapshotVersions>
  </versioning>
</metadata>`

func TestParseReleaseMetadata(t *testing.T) {
	md, err := ParseReleaseMetadata(releaseXML)
	if err != nil {
		t.Fatal(err)
	}
	if md.Release != "3.1.0" || md.Latest != "3.1.0" {
		t.Errorf("latest/release = %q/%q", md.Latest, md.Release)
	}
	if len(md.Versions) != 2 || md.Versions[1] != "3.1.0" {
		t.Errorf("versions = %v", md.Versions)
	}
	if got := md.LastUpdatedTime(); got != "2026-01-27T20:49:46Z" {
		t.Errorf("LastUpdatedTime = %q", got)
	}
}

func TestParseSnapshotMetadata(t *testing.T) {
	md, err := ParseSnapshotMetadata(snapshotXML)
	if err != nil {
		t.Fatal(err)
	}
	if md.Version != "3.0.4-SNAPSHOT" || md.BuildNumber != "4" {
		t.Errorf("version/build = %q/%q", md.Version, md.BuildNumber)
	}
	if len(md.SnapshotVersions) != 1 || md.SnapshotVersions[0].Classifier != "sources" {
		t.Errorf("snapshotVersions = %v", md.SnapshotVersions)
	}
	if got := md.LastUpdatedTime(); got != "2026-06-02T19:33:58Z" {
		t.Errorf("LastUpdatedTime = %q", got)
	}
}

func TestLastUpdatedTimeFallback(t *testing.T) {
	md := ReleaseMetadata{LastUpdated: "not-a-date"}
	if got := md.LastUpdatedTime(); got != "not-a-date" {
		t.Errorf("expected raw fallback, got %q", got)
	}
}
