package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CrowdStrike/codestrike/internal/scm/github"
)

func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *github.Client) {
	server := httptest.NewServer(handler)
	client := github.NewWithHTTPClient(github.Config{
		Owner:   "testowner",
		Repo:    "testrepo",
		Token:   "test-token",
		BaseURL: server.URL,
	}, server.Client())
	return server, client
}

func TestGetPullRequestDiff(t *testing.T) {
	wantDiff := "diff --git a/file.go b/file.go\n--- a/file.go\n+++ b/file.go\n@@ -1 +1 @@\n-old\n+new\n"

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/testowner/testrepo/pulls/1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github.v3.diff" {
			t.Errorf("unexpected accept header: %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(wantDiff)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	})
	defer server.Close()

	diff, err := client.GetPullRequestDiff(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if diff != wantDiff {
		t.Errorf("got %q, want %q", diff, wantDiff)
	}
}

func TestGetPullRequestDiff_NotFound(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	_, err := client.GetPullRequestDiff(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for missing PR")
	}
}

func TestPullRequestExists(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"number": 1}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	})
	defer server.Close()

	exists, err := client.PullRequestExists(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected PR to exist")
	}
}

func TestPullRequestExists_NotFound(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	exists, err := client.PullRequestExists(context.Background(), 999)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected PR to not exist")
	}
}

func TestPublishComment(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/testowner/testrepo/issues/1/comments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	})
	defer server.Close()

	err := client.PublishComment(context.Background(), 1, "LGTM!")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublishComment_NotFound(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	err := client.PublishComment(context.Background(), 999, "comment")
	if err == nil {
		t.Fatal("expected error for missing PR")
	}
}

func TestPublishComment_ReportsGitHubAPIError(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		if _, err := w.Write([]byte(`{"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/issues/comments#create-an-issue-comment"}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	})
	defer server.Close()

	err := client.PublishComment(context.Background(), 1, "comment")
	if err == nil {
		t.Fatal("expected permission error")
	}

	want := "GitHub API returned 403 Forbidden: Resource not accessible by personal access token (https://docs.github.com/rest/issues/comments#create-an-issue-comment)"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}
