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
			want: review.PRReference{Owner: "CrowdStrike", Repo: "codestrike", Number: 42},
		},
		{
			name: "trailing slash",
			url:  "https://github.com/owner/repo/pull/7/",
			want: review.PRReference{Owner: "owner", Repo: "repo", Number: 7},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := review.ParsePRURL(tt.url)
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
