// Package ghhelp is a small GitHub Actions helper. For now it has a single
// function, FindJobFailures, which fetches the log for a job (identified by its
// github.com job URL) and prints the failing-test sections.
package ghhelp

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// timestampPrefix matches the leading "2026-06-02T19:33:58.123Z " that GitHub
// prepends to each job-log line.
var timestampPrefix = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z `)

// FindJobFailures parses a github.com job URL, downloads that job's log, and
// prints the failing-test sections to stdout.
//
// Accepted URL shapes (matching the Java version):
//
//	https://github.com/synadia-io/nats.java.v3/actions/runs/22879139630/job/66377777516
//	https://github.com/nats-io/nats.java/actions/runs/25391554069/job/74467105487?pr=1564
func FindJobFailures(token, jobURL string) error {
	org, repo, jobID, err := parseJobURL(jobURL)
	if err != nil {
		return err
	}
	log, err := getJobLog(token, org, repo, jobID)
	if err != nil {
		return err
	}
	processJobLog(log)
	return nil
}

// parseJobURL extracts org, repo and job id from a github.com job URL.
func parseJobURL(jobURL string) (org, repo, jobID string, err error) {
	const marker = "github.com/"
	at := strings.Index(jobURL, marker)
	if at == -1 {
		return "", "", "", fmt.Errorf("not a github.com url: %s", jobURL)
	}
	rest := jobURL[at+len(marker):]
	if q := strings.LastIndex(rest, "?"); q != -1 {
		rest = rest[:q]
	}

	slash := strings.Index(rest, "/")
	if slash == -1 {
		return "", "", "", fmt.Errorf("could not find org in url: %s", jobURL)
	}
	org = rest[:slash]
	rest = rest[slash+1:]

	slash = strings.Index(rest, "/")
	if slash == -1 {
		return "", "", "", fmt.Errorf("could not find repo in url: %s", jobURL)
	}
	repo = rest[:slash]

	jobMarker := strings.Index(rest, "job/")
	if jobMarker == -1 {
		return "", "", "", fmt.Errorf("could not find job id in url: %s", jobURL)
	}
	jobID = rest[jobMarker+len("job/"):]
	if jobID == "" {
		return "", "", "", fmt.Errorf("empty job id in url: %s", jobURL)
	}
	return org, repo, jobID, nil
}

// getJobLog downloads the raw log for a job. The GitHub API responds with a
// redirect to a signed URL; net/http follows it by default, but the Authorization
// header must not be forwarded to the storage host, so we only set it on the
// initial api.github.com request.
func getJobLog(token, org, repo, jobID string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/jobs/%s/logs", org, repo, jobID)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Drop the Authorization header when redirected off api.github.com.
			if req.URL.Host != "api.github.com" {
				req.Header.Del("Authorization")
			}
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

// processJobLog prints the lines of each failing test, from the "() FAILED"
// marker until the next "() STARTED" or test-summary line.
func processJobLog(log string) {
	printing := false
	for _, line := range strings.Split(log, "\n") {
		clean := timestampPrefix.ReplaceAllString(line, "")
		switch {
		case strings.Contains(line, "() FAILED"):
			printing = true
			fmt.Println(clean)
		case printing:
			if strings.Contains(line, "() STARTED") || strings.Contains(line, " tests completed, ") {
				printing = false
			} else {
				fmt.Println(clean)
			}
		}
	}
}
