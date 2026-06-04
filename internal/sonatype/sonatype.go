// Package sonatype is a Go port of the Java SonatypeChecker. It reads a list of
// projects (CheckerParams) and, for each requested version, fetches the Maven
// metadata from the release and snapshot repositories and reports when it was
// last updated.
package sonatype

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	releaseLabel            = "artifactId:version"
	snapshotLabel           = "artifactId:version-SNAPSHOT"
	releaseMetadataTemplate = "https://repo1.maven.org/maven2/groupId/artifactId/maven-metadata.xml"
	// version-SNAPSHOT in the template doubles as the marker for snapshot handling.
	snapshotMetadataTemplate = "https://central.sonatype.com/repository/maven-snapshots/groupId/artifactId/version-SNAPSHOT/maven-metadata.xml"
)

// JDK qualifier sets, mirroring the Java constants. A nil entry means "no
// qualifier" (the bare artifact with no -jdkNN suffix).
var (
	JDKS17On    = []*string{strp("17"), strp("21"), strp("25")}
	JDKS8On     = []*string{nil, strp("17"), strp("21"), strp("25")}
	NoQualifier = []*string{nil}
)

func strp(s string) *string { return &s }

// Debug, when true, prints the URL being checked before each request.
var Debug = false

var httpClient = &http.Client{Timeout: 5 * time.Second}

// CheckerParams describes one project to check. JSON property names use
// underscores to match the shared .sonatype-checker.json config file.
type CheckerParams struct {
	Project          string   `json:"project"`
	GroupID          string   `json:"group_id"`
	Component        string   `json:"component"`
	JdkQualifiers    []string `json:"jdk_qualifiers"`
	ReleaseVersions  []string `json:"release_versions"`
	SnapshotVersions []string `json:"snapshot_versions"`
}

// LoadParams reads a JSON array of CheckerParams from the given file.
func LoadParams(filename string) ([]CheckerParams, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var params []CheckerParams
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}
	return params, nil
}

// Check runs the release and snapshot checks for a single project, printing the
// results to stdout (matching the Java output format).
func Check(p CheckerParams) error {
	fmt.Printf("\n%s %s:%s\n", p.Project, p.GroupID, p.Component)
	fmt.Println("  RELEASES")
	for _, vr := range p.ReleaseVersions {
		if err := check(p.GroupID, p.Component, p.JdkQualifiers, vr, releaseLabel, releaseMetadataTemplate); err != nil {
			return err
		}
	}
	fmt.Println("  SNAPSHOTS")
	for _, vs := range p.SnapshotVersions {
		if err := check(p.GroupID, p.Component, p.JdkQualifiers, vs, snapshotLabel, snapshotMetadataTemplate); err != nil {
			return err
		}
	}
	return nil
}

func check(groupID, component string, jdkQualifiers []string, version, labelTemplate, metaTemplate string) error {
	// jdkQualifiers from JSON is a []string where an empty/absent string means
	// "no qualifier" (the JSON encodes that as null, decoded to "").
	quals := jdkQualifiers
	if len(quals) == 0 {
		quals = []string{""}
	}
	for _, j := range quals {
		artifactID := component
		if j != "" {
			artifactID = component + "-jdk" + j
		}
		ident := strings.NewReplacer("artifactId", artifactID, "version", version).Replace(labelTemplate)
		url := strings.NewReplacer(
			"groupId", strings.ReplaceAll(groupID, ".", "/"),
			"artifactId", artifactID,
			"version", version,
		).Replace(metaTemplate)

		if Debug {
			fmt.Printf("    ? Checking %s for %s\n", url, labelTemplate)
		}

		body, err := ReadURL(url)
		if err != nil {
			return err
		}
		if body == "" {
			fmt.Printf("    %s\n    | Not Found\n    | Url: %s\n", ident, url)
			continue
		}

		if strings.Contains(metaTemplate, "-SNAPSHOT") {
			meta, err := ParseSnapshotMetadata(body)
			if err != nil {
				return err
			}
			fmt.Printf("    %s\n    | Last Updated: %s\n    | Url: %s\n", ident, meta.LastUpdatedTime(), url)
		} else {
			meta, err := ParseReleaseMetadata(body)
			if err != nil {
				return err
			}
			fmt.Printf("    %s\n    | Last Updated: %s\n    | Url: %s\n", ident, meta.LastUpdatedTime(), url)
		}
	}
	return nil
}

// ReadURL performs a GET and returns the response body, or "" if the file does
// not exist (any non-200 status).
func ReadURL(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ReleaseMetadata mirrors a release maven-metadata.xml.
type ReleaseMetadata struct {
	GroupID     string   `xml:"groupId"`
	ArtifactID  string   `xml:"artifactId"`
	Latest      string   `xml:"versioning>latest"`
	Release     string   `xml:"versioning>release"`
	LastUpdated string   `xml:"versioning>lastUpdated"`
	Versions    []string `xml:"versioning>versions>version"`
}

// LastUpdatedTime parses LastUpdated (yyyyMMddHHmmss, UTC) into a formatted
// timestamp, or returns the raw value if it cannot be parsed.
func (m ReleaseMetadata) LastUpdatedTime() string { return formatLastUpdated(m.LastUpdated) }

// SnapshotMetadata mirrors a snapshot maven-metadata.xml.
type SnapshotMetadata struct {
	GroupID          string            `xml:"groupId"`
	ArtifactID       string            `xml:"artifactId"`
	Version          string            `xml:"version"`
	Timestamp        string            `xml:"versioning>snapshot>timestamp"`
	BuildNumber      string            `xml:"versioning>snapshot>buildNumber"`
	LastUpdated      string            `xml:"versioning>lastUpdated"`
	SnapshotVersions []SnapshotVersion `xml:"versioning>snapshotVersions>snapshotVersion"`
}

// SnapshotVersion is one entry under snapshotVersions.
type SnapshotVersion struct {
	Classifier string `xml:"classifier"`
	Extension  string `xml:"extension"`
	Value      string `xml:"value"`
	Updated    string `xml:"updated"`
}

// LastUpdatedTime parses LastUpdated (yyyyMMddHHmmss, UTC) into a formatted
// timestamp, or returns the raw value if it cannot be parsed.
func (m SnapshotMetadata) LastUpdatedTime() string { return formatLastUpdated(m.LastUpdated) }

// ParseReleaseMetadata parses a release maven-metadata.xml document.
func ParseReleaseMetadata(content string) (ReleaseMetadata, error) {
	var md ReleaseMetadata
	if content == "" {
		return md, nil
	}
	err := xml.Unmarshal([]byte(content), &md)
	return md, err
}

// ParseSnapshotMetadata parses a snapshot maven-metadata.xml document.
func ParseSnapshotMetadata(content string) (SnapshotMetadata, error) {
	var md SnapshotMetadata
	if content == "" {
		return md, nil
	}
	err := xml.Unmarshal([]byte(content), &md)
	return md, err
}

// Maven metadata lastUpdated is yyyyMMddHHmmss in UTC.
const lastUpdatedLayout = "20060102150405"

func formatLastUpdated(value string) string {
	if value == "" {
		return ""
	}
	t, err := time.ParseInLocation(lastUpdatedLayout, value, time.UTC)
	if err != nil {
		return value
	}
	// RFC3339 in UTC renders like the Java ZonedDateTime output, e.g. 2026-02-25T20:45:37Z.
	return t.Format(time.RFC3339)
}
