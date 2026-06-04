package ghhelp

import "testing"

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
