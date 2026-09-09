package bitbucket_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CrowdStrike/codestrike/internal/scm/bitbucket"
)

func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *bitbucket.Client) {
	server := httptest.NewServer(handler)
	client := bitbucket.NewWithHTTPClient(bitbucket.Config{
		Workspace: "testworkspace",
		RepoSlug:  "testrepo",
		Token:     "test-token",
		BaseURL:   server.URL,
	}, server.Client())
	return server, client
}

func TestPullRequestExists(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/testworkspace/testrepo/pullrequests/1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 1}`))
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
	server, client := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
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

func TestGetPullRequestDiff(t *testing.T) {
	wantDiff := "diff --git a/file.go b/file.go\n--- a/file.go\n+++ b/file.go\n@@ -1 +1 @@\n-old\n+new\n"

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/testworkspace/testrepo/pullrequests/1/diff" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(wantDiff))
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
	server, client := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	_, err := client.GetPullRequestDiff(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for missing PR")
	}
}

func TestGetPullRequestFiles(t *testing.T) {
	diffstatResp := `{
		"values": [
			{"status": "modified", "new": {"path": "file.go"}, "old": {"path": "file.go"}},
			{"status": "added", "new": {"path": "new.go"}, "old": {}}
		]
	}`
	fullDiff := "diff --git a/file.go b/file.go\n--- a/file.go\n+++ b/file.go\n@@ -1 +1 @@\n-old\n+new\n" +
		"diff --git a/new.go b/new.go\n--- /dev/null\n+++ b/new.go\n@@ -0,0 +1 @@\n+package main\n"

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/diffstat"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(diffstatResp))
		case strings.HasSuffix(r.URL.Path, "/diff"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fullDiff))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	})
	defer server.Close()

	files, err := client.GetPullRequestFiles(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	if files[0].Filename != "file.go" {
		t.Errorf("expected filename %q, got %q", "file.go", files[0].Filename)
	}
	if files[0].Status != "modified" {
		t.Errorf("expected status %q, got %q", "modified", files[0].Status)
	}
	if files[0].Patch == "" {
		t.Error("expected non-empty patch for file.go")
	}

	if files[1].Filename != "new.go" {
		t.Errorf("expected filename %q, got %q", "new.go", files[1].Filename)
	}
	if files[1].Status != "added" {
		t.Errorf("expected status %q, got %q", "added", files[1].Status)
	}
}

func TestGetFileContent(t *testing.T) {
	wantContent := "package main\n\nfunc main() {}\n"

	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/testworkspace/testrepo/src/abc123/main.go" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(wantContent))
	})
	defer server.Close()

	content, err := client.GetFileContent(context.Background(), "main.go", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if content != wantContent {
		t.Errorf("got %q, want %q", content, wantContent)
	}
}

func TestGetFileContent_NotFound(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	_, err := client.GetFileContent(context.Background(), "missing.go", "abc123")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPublishComment(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/testworkspace/testrepo/pullrequests/1/comments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var body struct {
			Content struct {
				Raw string `json:"raw"`
			} `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if body.Content.Raw != "LGTM!" {
			t.Errorf("expected comment %q, got %q", "LGTM!", body.Content.Raw)
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
	server, client := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	err := client.PublishComment(context.Background(), 999, "comment")
	if err == nil {
		t.Fatal("expected error for missing PR")
	}
}

func TestPublishComment_ReportsAPIError(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"Insufficient permissions"}}`))
	})
	defer server.Close()

	err := client.PublishComment(context.Background(), 1, "comment")
	if err == nil {
		t.Fatal("expected permission error")
	}
	if !strings.Contains(err.Error(), "Insufficient permissions") {
		t.Errorf("error = %q, want it to contain %q", err, "Insufficient permissions")
	}
}

func TestGetPRComments(t *testing.T) {
	commentsResp := `{
		"values": [
			{"id": 1, "content": {"raw": "general comment"}, "user": {"display_name": "Alice"}, "created_on": "2024-01-15T10:00:00Z", "inline": null},
			{"id": 2, "content": {"raw": "inline comment"}, "user": {"display_name": "Bob"}, "created_on": "2024-01-15T11:00:00Z", "inline": {"path": "file.go", "to": 42}}
		]
	}`

	server, client := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(commentsResp))
	})
	defer server.Close()

	comments, err := client.GetPRComments(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 general comment, got %d", len(comments))
	}
	if comments[0].Author != "Alice" {
		t.Errorf("expected author %q, got %q", "Alice", comments[0].Author)
	}
	if comments[0].Body != "general comment" {
		t.Errorf("expected body %q, got %q", "general comment", comments[0].Body)
	}
}

func TestGetPRReviewComments(t *testing.T) {
	commentsResp := `{
		"values": [
			{"id": 1, "content": {"raw": "general comment"}, "user": {"display_name": "Alice"}, "created_on": "2024-01-15T10:00:00Z", "inline": null},
			{"id": 2, "content": {"raw": "inline comment"}, "user": {"display_name": "Bob"}, "created_on": "2024-01-15T11:00:00Z", "inline": {"path": "file.go", "to": 42}},
			{"id": 3, "content": {"raw": "reply"}, "user": {"display_name": "Carol"}, "created_on": "2024-01-15T12:00:00Z", "inline": {"path": "file.go", "to": 42}, "parent": {"id": 2}}
		]
	}`

	server, client := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(commentsResp))
	})
	defer server.Close()

	comments, err := client.GetPRReviewComments(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 2 {
		t.Fatalf("expected 2 inline comments, got %d", len(comments))
	}
	if comments[0].Path != "file.go" {
		t.Errorf("expected path %q, got %q", "file.go", comments[0].Path)
	}
	if comments[0].Line != 42 {
		t.Errorf("expected line 42, got %d", comments[0].Line)
	}
	if comments[1].InReplyTo != 2 {
		t.Errorf("expected InReplyTo 2, got %d", comments[1].InReplyTo)
	}
}
