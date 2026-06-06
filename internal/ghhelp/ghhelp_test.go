package ghhelp

import (
	"strings"
	"testing"
)

func TestParseJobURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		org     string
		repo    string
		jobID   string
		wantErr bool
	}{
		{
			name:  "plain",
			url:   "https://github.com/synadia-io/nats.java.v3/actions/runs/22879139630/job/66377777516",
			org:   "synadia-io",
			repo:  "nats.java.v3",
			jobID: "66377777516",
		},
		{
			name:  "with pr query",
			url:   "https://github.com/nats-io/nats.java/actions/runs/25391554069/job/74467105487?pr=1564",
			org:   "nats-io",
			repo:  "nats.java",
			jobID: "74467105487",
		},
		{
			name:    "not a github url",
			url:     "https://example.com/foo/bar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, repo, jobID, err := parseJobURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if org != tt.org || repo != tt.repo || jobID != tt.jobID {
				t.Errorf("got (%q,%q,%q), want (%q,%q,%q)", org, repo, jobID, tt.org, tt.repo, tt.jobID)
			}
		})
	}
}

// sampleLog has three failing tests: two share a reason, one is unique. It also
// includes timestamp prefixes and STARTED / "tests completed," terminators.
const sampleLog = `2026-06-06T10:00:00.000Z AlphaTests > testOne() STARTED
2026-06-06T10:00:01.000Z AlphaTests > testOne() FAILED
2026-06-06T10:00:01.001Z     org.opentest4j.AssertionFailedError: boom
2026-06-06T10:00:01.002Z         at app//Alpha.java:10
2026-06-06T10:00:02.000Z AlphaTests > testTwo() STARTED
2026-06-06T10:00:03.000Z AlphaTests > testTwo() FAILED
2026-06-06T10:00:03.001Z     org.opentest4j.AssertionFailedError: boom
2026-06-06T10:00:03.002Z         at app//Alpha.java:20
2026-06-06T10:00:04.000Z BetaTests > testThree() FAILED
2026-06-06T10:00:04.001Z     java.lang.NullPointerException: npe
2026-06-06T10:00:05.000Z 42 tests completed, 3 failed`

func TestParseFailures(t *testing.T) {
	got := parseFailures(sampleLog)
	if len(got) != 3 {
		t.Fatalf("got %d failures, want 3", len(got))
	}
	want := []struct{ test, reason string }{
		{"AlphaTests > testOne()", "org.opentest4j.AssertionFailedError: boom"},
		{"AlphaTests > testTwo()", "org.opentest4j.AssertionFailedError: boom"},
		{"BetaTests > testThree()", "java.lang.NullPointerException: npe"},
	}
	for i, w := range want {
		if got[i].Test != w.test || got[i].Reason != w.reason {
			t.Errorf("failure %d: got (%q,%q), want (%q,%q)", i, got[i].Test, got[i].Reason, w.test, w.reason)
		}
	}
	// The timestamp prefix must be stripped from body lines.
	if strings.Contains(got[0].Body[0], "2026-") {
		t.Errorf("timestamp not stripped from body: %q", got[0].Body[0])
	}
}

func TestSummarize(t *testing.T) {
	var b strings.Builder
	summarize(&b, parseFailures(sampleLog))
	out := b.String()

	if !strings.Contains(out, "3 failing test(s), 2 distinct failure(s):") {
		t.Errorf("missing header line; got:\n%s", out)
	}
	// The shared reason (2 tests) must be grouped under one [2x] heading listing
	// both tests, and appear before the single-test group.
	shared := strings.Index(out, "[2x] org.opentest4j.AssertionFailedError: boom")
	single := strings.Index(out, "[1x] java.lang.NullPointerException: npe")
	if shared == -1 || single == -1 {
		t.Fatalf("missing grouped headings; got:\n%s", out)
	}
	if shared > single {
		t.Errorf("most-shared reason should come first; got:\n%s", out)
	}
	for _, test := range []string{"AlphaTests > testOne()", "AlphaTests > testTwo()", "BetaTests > testThree()"} {
		if !strings.Contains(out, test) {
			t.Errorf("output missing test %q; got:\n%s", test, out)
		}
	}
}
