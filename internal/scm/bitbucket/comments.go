package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CrowdStrike/codestrike/internal/scm"
)

// GetPRComments returns general (non-inline) comments on the PR.
func (c *Client) GetPRComments(ctx context.Context, number int) ([]scm.PRComment, error) {
	all, err := c.fetchComments(ctx, number)
	if err != nil {
		return nil, err
	}

	var general []scm.PRComment
	for _, comment := range all {
		if comment.Path == "" {
			general = append(general, comment)
		}
	}

	return general, nil
}

// GetPRReviewComments returns inline (file-level) comments on the PR.
func (c *Client) GetPRReviewComments(ctx context.Context, number int) ([]scm.PRComment, error) {
	all, err := c.fetchComments(ctx, number)
	if err != nil {
		return nil, err
	}

	var inline []scm.PRComment
	for _, comment := range all {
		if comment.Path != "" {
			inline = append(inline, comment)
		}
	}

	return inline, nil
}

func (c *Client) fetchComments(ctx context.Context, number int) ([]scm.PRComment, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/comments",
		c.config.BaseURL, c.config.Workspace, c.config.RepoSlug, number)

	var allComments []scm.PRComment
	for url != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		c.setAuth(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("pull request #%d not found", number)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}

		var page struct {
			Values []bbComment `json:"values"`
			Next   string      `json:"next"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		resp.Body.Close()

		for _, bc := range page.Values {
			comment := scm.PRComment{
				ID:        bc.ID,
				Author:    bc.User.DisplayName,
				Body:      bc.Content.Raw,
				CreatedAt: bc.CreatedOn,
			}
			if bc.Inline != nil {
				comment.Path = bc.Inline.Path
				comment.Line = bc.Inline.To
			}
			if bc.Parent != nil {
				comment.InReplyTo = bc.Parent.ID
			}
			allComments = append(allComments, comment)
		}

		url = page.Next
	}

	return allComments, nil
}

type bbComment struct {
	ID        int64     `json:"id"`
	CreatedOn time.Time `json:"created_on"`
	Content   struct {
		Raw string `json:"raw"`
	} `json:"content"`
	User struct {
		DisplayName string `json:"display_name"`
	} `json:"user"`
	Inline *struct {
		Path string `json:"path"`
		To   int    `json:"to"`
	} `json:"inline"`
	Parent *struct {
		ID int64 `json:"id"`
	} `json:"parent"`
}
