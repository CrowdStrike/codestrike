package review

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/CrowdStrike/codestrike/internal/scm"
)

type mockClient struct {
	generalComments []scm.PRComment
	reviewComments  []scm.PRComment
}

func (m *mockClient) GetPRComments(_ context.Context, _ int) ([]scm.PRComment, error) {
	return m.generalComments, nil
}

func (m *mockClient) GetPRReviewComments(_ context.Context, _ int) ([]scm.PRComment, error) {
	return m.reviewComments, nil
}

func (m *mockClient) GetPullRequestDiff(_ context.Context, _ int) (string, error) { return "", nil }
func (m *mockClient) GetPullRequestFiles(_ context.Context, _ int) ([]scm.PullRequestFile, error) {
	return nil, nil
}
func (m *mockClient) GetFileContent(_ context.Context, _, _ string) (string, error) { return "", nil }
func (m *mockClient) PullRequestExists(_ context.Context, _ int) (bool, error)     { return true, nil }
func (m *mockClient) PublishComment(_ context.Context, _ int, _ string) error       { return nil }

func newTestPipeline(client scm.Client) *Pipeline {
	log := zerolog.Nop()
	return &Pipeline{
		client: client,
		logger: &log,
	}
}

func TestFetchExistingComments_SeparatesTrustedAndUntrusted(t *testing.T) {
	now := time.Now()
	client := &mockClient{
		generalComments: []scm.PRComment{
			{
				ID:        1,
				Author:    "codestrike[bot]",
				Body:      "<!-- codestrike:review -->\n## Code Review\n\n**a.go:10**\nmissing error check\n",
				CreatedAt: now.Add(-2 * time.Hour),
			},
			{
				ID:        2,
				Author:    "dev-user",
				Body:      "Good catch, I'll fix the error handling.",
				CreatedAt: now.Add(-1 * time.Hour),
			},
			{
				ID:        3,
				Author:    "another-user",
				Body:      "LGTM overall",
				CreatedAt: now.Add(-30 * time.Minute),
			},
		},
	}

	p := newTestPipeline(client)
	ownComments, userFeedback := p.fetchExistingCommentsContext(context.Background(), 1)

	if !strings.Contains(ownComments, "missing error check") {
		t.Error("expected own comments to contain codestrike's review")
	}
	if strings.Contains(ownComments, "Good catch") {
		t.Error("expected own comments to NOT contain user feedback")
	}

	if !strings.Contains(userFeedback, "Good catch") {
		t.Error("expected user feedback to contain reply after codestrike review")
	}
	if !strings.Contains(userFeedback, "dev-user") {
		t.Error("expected user feedback to contain author name")
	}
	if !strings.Contains(userFeedback, "LGTM overall") {
		t.Error("expected user feedback to include other comments after codestrike review")
	}
}

func TestFetchExistingComments_ExcludesCommentsBeforeReview(t *testing.T) {
	now := time.Now()
	client := &mockClient{
		generalComments: []scm.PRComment{
			{
				ID:        1,
				Author:    "dev-user",
				Body:      "pre-review discussion comment",
				CreatedAt: now.Add(-3 * time.Hour),
			},
			{
				ID:        2,
				Author:    "codestrike[bot]",
				Body:      "<!-- codestrike:review -->\n## Code Review\n\nfindings here",
				CreatedAt: now.Add(-2 * time.Hour),
			},
			{
				ID:        3,
				Author:    "dev-user",
				Body:      "post-review feedback",
				CreatedAt: now.Add(-1 * time.Hour),
			},
		},
	}

	p := newTestPipeline(client)
	_, userFeedback := p.fetchExistingCommentsContext(context.Background(), 1)

	if strings.Contains(userFeedback, "pre-review discussion") {
		t.Error("expected user feedback to NOT include comments posted before codestrike's review")
	}
	if !strings.Contains(userFeedback, "post-review feedback") {
		t.Error("expected user feedback to include comments posted after codestrike's review")
	}
}

func TestFetchExistingComments_CapsAt10NewestFirst(t *testing.T) {
	now := time.Now()
	codestrikeReview := scm.PRComment{
		ID:        1,
		Author:    "codestrike[bot]",
		Body:      "<!-- codestrike:review -->\nreview content",
		CreatedAt: now.Add(-24 * time.Hour),
	}

	var userComments []scm.PRComment
	for i := range 15 {
		userComments = append(userComments, scm.PRComment{
			ID:        int64(i + 10),
			Author:    "user",
			Body:      strings.Repeat("x", 10),
			CreatedAt: now.Add(-time.Duration(15-i) * time.Hour),
		})
	}

	client := &mockClient{
		generalComments: append([]scm.PRComment{codestrikeReview}, userComments...),
	}

	p := newTestPipeline(client)
	_, userFeedback := p.fetchExistingCommentsContext(context.Background(), 1)

	lines := strings.Split(strings.TrimSpace(userFeedback), "\n")
	if len(lines) > 10 {
		t.Errorf("expected at most 10 feedback entries, got %d", len(lines))
	}
}

func TestFetchExistingComments_NoCodestrikeReview_NoFeedback(t *testing.T) {
	client := &mockClient{
		generalComments: []scm.PRComment{
			{
				ID:        1,
				Author:    "user",
				Body:      "some random comment",
				CreatedAt: time.Now(),
			},
		},
	}

	p := newTestPipeline(client)
	ownComments, userFeedback := p.fetchExistingCommentsContext(context.Background(), 1)

	if ownComments != "" {
		t.Errorf("expected empty own comments, got %q", ownComments)
	}
	if userFeedback != "" {
		t.Errorf("expected empty user feedback when no codestrike review exists, got %q", userFeedback)
	}
}

func TestFetchExistingComments_MultipleRounds(t *testing.T) {
	now := time.Now()
	client := &mockClient{
		generalComments: []scm.PRComment{
			{
				ID:        1,
				Author:    "codestrike[bot]",
				Body:      "<!-- codestrike:review -->\nfirst review",
				CreatedAt: now.Add(-4 * time.Hour),
			},
			{
				ID:        2,
				Author:    "user",
				Body:      "old feedback to first review",
				CreatedAt: now.Add(-3 * time.Hour),
			},
			{
				ID:        3,
				Author:    "codestrike[bot]",
				Body:      "<!-- codestrike:review -->\nsecond review",
				CreatedAt: now.Add(-2 * time.Hour),
			},
			{
				ID:        4,
				Author:    "user",
				Body:      "recent feedback to second review",
				CreatedAt: now.Add(-1 * time.Hour),
			},
		},
	}

	p := newTestPipeline(client)
	ownComments, userFeedback := p.fetchExistingCommentsContext(context.Background(), 1)

	if !strings.Contains(ownComments, "first review") {
		t.Error("expected own comments to contain first review (for dedup)")
	}
	if !strings.Contains(ownComments, "second review") {
		t.Error("expected own comments to contain second review (for dedup)")
	}

	if strings.Contains(userFeedback, "old feedback") {
		t.Error("expected user feedback to NOT include feedback from before the latest review")
	}
	if !strings.Contains(userFeedback, "recent feedback") {
		t.Error("expected user feedback to include feedback after the latest review")
	}
}
