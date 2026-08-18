package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CrowdStrike/codestrike/internal/scm"
)

func (c *Client) GetPRComments(ctx context.Context, number int) ([]scm.PRComment, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.config.BaseURL, c.config.Owner, c.config.Repo, number)
	return c.fetchComments(ctx, url)
}

func (c *Client) GetPRReviewComments(ctx context.Context, number int) ([]scm.PRComment, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/comments", c.config.BaseURL, c.config.Owner, c.config.Repo, number)
	return c.fetchComments(ctx, url)
}

func (c *Client) fetchComments(ctx context.Context, url string) ([]scm.PRComment, error) {
	var allComments []scm.PRComment
	page := 1

	for {
		pageURL := fmt.Sprintf("%s?per_page=100&page=%d", url, page)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		c.setAuth(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("resource not found")
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}

		var ghComments []ghComment
		if err := json.NewDecoder(resp.Body).Decode(&ghComments); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		resp.Body.Close()

		if len(ghComments) == 0 {
			break
		}

		for _, gc := range ghComments {
			allComments = append(allComments, scm.PRComment{
				ID:        gc.ID,
				Author:    gc.User.Login,
				Body:      gc.Body,
				CreatedAt: gc.CreatedAt,
				Path:      gc.Path,
				Line:      gc.Line,
				InReplyTo: gc.InReplyToID,
			})
		}

		page++
		if len(ghComments) < 100 {
			break
		}
	}

	return allComments, nil
}

type ghComment struct {
	ID          int64     `json:"id"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	Path        string    `json:"path"`
	Line        int       `json:"line"`
	InReplyToID int64     `json:"in_reply_to_id"`
	User        struct {
		Login string `json:"login"`
	} `json:"user"`
}
