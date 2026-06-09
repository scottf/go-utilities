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

	// Three distinct tests, none combined — even though testOne and testTwo
	// share the same reason ("boom"), they are different tests.
	if !strings.Contains(out, "3 failing test(s), 3 distinct:") {
		t.Errorf("missing/wrong header line; got:\n%s", out)
	}
	for _, header := range []string{
		"AlphaTests > testOne() FAILED",
		"AlphaTests > testTwo() FAILED",
		"BetaTests > testThree() FAILED",
	} {
		if !strings.Contains(out, header) {
			t.Errorf("output missing failure header %q; got:\n%s", header, out)
		}
	}
	// The stack trace must be shown (not just the reason line).
	if !strings.Contains(out, "at app//Alpha.java:10") {
		t.Errorf("output missing stack trace; got:\n%s", out)
	}
	// No (xN) count when nothing repeats.
	if strings.Contains(out, "(x") {
		t.Errorf("unexpected duplicate count for non-repeating tests; got:\n%s", out)
	}
}

// dupLog has the same test (AlphaTests.testOne) failing twice with an identical
// stack trace, plus one other test.
const dupLog = `2026-06-06T10:00:00.000Z AlphaTests > testOne() FAILED
2026-06-06T10:00:00.001Z     org.opentest4j.AssertionFailedError: boom
2026-06-06T10:00:00.002Z         at app//Alpha.java:10
2026-06-06T10:00:01.000Z AlphaTests > testOne() STARTED
2026-06-06T10:00:02.000Z AlphaTests > testOne() FAILED
2026-06-06T10:00:02.001Z     org.opentest4j.AssertionFailedError: boom
2026-06-06T10:00:02.002Z         at app//Alpha.java:10
2026-06-06T10:00:03.000Z BetaTests > testThree() FAILED
2026-06-06T10:00:03.001Z     java.lang.NullPointerException: npe
2026-06-06T10:00:04.000Z 10 tests completed, 3 failed`

func TestSummarizeCollapsesRepeats(t *testing.T) {
	var b strings.Builder
	summarize(&b, parseFailures(dupLog))
	out := b.String()

	if !strings.Contains(out, "3 failing test(s), 2 distinct:") {
		t.Errorf("missing/wrong header line; got:\n%s", out)
	}
	// The repeated test is shown once with a count.
	if !strings.Contains(out, "AlphaTests > testOne() FAILED   (x2)") {
		t.Errorf("repeated test not collapsed with count; got:\n%s", out)
	}
	if strings.Count(out, "at app//Alpha.java:10") != 1 {
		t.Errorf("repeated stack trace should appear exactly once; got:\n%s", out)
	}
	// The non-repeating test has no count.
	if !strings.Contains(out, "BetaTests > testThree() FAILED\n") {
		t.Errorf("missing non-repeating test header; got:\n%s", out)
	}
}
