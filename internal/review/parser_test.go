package review_test

import (
	"testing"

	"github.com/CrowdStrike/codestrike/internal/review"
)

func TestParsePRURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    review.PRReference
		wantErr bool
	}{
		{
			name: "standard github URL",
			url:  "https://github.com/CrowdStrike/codestrike/pull/42",
			want: review.PRReference{Provider: review.ProviderGitHub, Owner: "CrowdStrike", Repo: "codestrike", Number: 42},
		},
		{
			name: "github trailing slash",
			url:  "https://github.com/owner/repo/pull/7/",
			want: review.PRReference{Provider: review.ProviderGitHub, Owner: "owner", Repo: "repo", Number: 7},
		},
		{
			name: "standard bitbucket URL",
			url:  "https://bitbucket.org/workspace/my-repo/pull-requests/123",
			want: review.PRReference{Provider: review.ProviderBitbucket, Owner: "workspace", Repo: "my-repo", Number: 123},
		},
		{
			name: "bitbucket trailing slash",
			url:  "https://bitbucket.org/workspace/repo/pull-requests/5/",
			want: review.PRReference{Provider: review.ProviderBitbucket, Owner: "workspace", Repo: "repo", Number: 5},
		},
		{
			name:    "invalid URL missing pull segment",
			url:     "https://github.com/owner/repo/issues/1",
			wantErr: true,
		},
		{
			name:    "invalid PR number",
			url:     "https://github.com/owner/repo/pull/abc",
			wantErr: true,
		},
		{
			name:    "too short URL",
			url:     "https://github.com/pull/1",
			wantErr: true,
		},
		{
			name:    "invalid bitbucket PR number",
			url:     "https://bitbucket.org/workspace/repo/pull-requests/abc",
			wantErr: true,
		},
		{
			name:    "too short bitbucket URL",
			url:     "https://bitbucket.org/pull-requests/1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := review.ParsePrUrl(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
