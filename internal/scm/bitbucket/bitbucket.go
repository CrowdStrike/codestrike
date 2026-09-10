package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/CrowdStrike/codestrike/internal/scm"
)

var _ scm.Client = (*Client)(nil)

// Config holds the Bitbucket API connection settings.
type Config struct {
	Workspace string
	RepoSlug  string
	Token     string
	BaseURL   string
}

// Client implements scm.Client for Bitbucket Cloud.
type Client struct {
	config     Config
	httpClient *http.Client
}

// New creates a Bitbucket client with default HTTP client.
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.bitbucket.org/2.0"
	}
	return &Client{
		config:     cfg,
		httpClient: http.DefaultClient,
	}
}

// NewWithHTTPClient creates a Bitbucket client with a custom HTTP client (for testing).
func NewWithHTTPClient(cfg Config, httpClient *http.Client) *Client {
	c := New(cfg)
	c.httpClient = httpClient
	return c
}

func (c *Client) PullRequestExists(ctx context.Context, number int) (bool, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d",
		c.config.BaseURL, c.config.Workspace, c.config.RepoSlug, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return true, nil
}

func (c *Client) GetPullRequestDescription(ctx context.Context, number int) (string, string, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d",
		c.config.BaseURL, c.config.Workspace, c.config.RepoSlug, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var pr struct {
		Title   string `json:"title"`
		Summary struct {
			Raw string `json:"raw"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return "", "", fmt.Errorf("decoding response: %w", err)
	}

	return pr.Title, pr.Summary.Raw, nil
}

func (c *Client) GetPullRequestDiff(ctx context.Context, number int) (string, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/diff",
		c.config.BaseURL, c.config.Workspace, c.config.RepoSlug, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("pull request #%d not found", number)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return string(body), nil
}

func (c *Client) GetPullRequestFiles(ctx context.Context, number int) ([]scm.PullRequestFile, error) {
	entries, err := c.fetchDiffstat(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("fetching diffstat: %w", err)
	}

	fullDiff, err := c.GetPullRequestDiff(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("fetching diff: %w", err)
	}
	patches := splitDiff(fullDiff)

	files := make([]scm.PullRequestFile, 0, len(entries))
	for _, e := range entries {
		files = append(files, scm.PullRequestFile{
			Filename: e.Path,
			Status:   e.Status,
			Patch:    patches[e.Path],
		})
	}

	return files, nil
}

func (c *Client) GetFileContent(ctx context.Context, path, ref string) (string, error) {
	if ref == "" {
		ref = "HEAD"
	}
	url := fmt.Sprintf("%s/repositories/%s/%s/src/%s/%s",
		c.config.BaseURL, c.config.Workspace, c.config.RepoSlug, ref, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("file %q not found", path)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return string(body), nil
}

func (c *Client) PublishComment(ctx context.Context, number int, body string) error {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/comments",
		c.config.BaseURL, c.config.Workspace, c.config.RepoSlug, number)

	payload, err := json.Marshal(map[string]any{
		"content": map[string]string{"raw": body},
	})
	if err != nil {
		return fmt.Errorf("marshaling comment: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("pull request #%d not found", number)
	}
	if resp.StatusCode != http.StatusCreated {
		return bitbucketAPIError(resp)
	}

	return nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}
}

type diffstatEntry struct {
	Status string `json:"status"`
	Path   string `json:"-"`
}

type diffstatNewOld struct {
	Path string `json:"path"`
}

type diffstatRaw struct {
	Status string         `json:"status"`
	New    diffstatNewOld `json:"new"`
	Old    diffstatNewOld `json:"old"`
}

func (c *Client) fetchDiffstat(ctx context.Context, number int) ([]diffstatEntry, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/diffstat",
		c.config.BaseURL, c.config.Workspace, c.config.RepoSlug, number)

	var entries []diffstatEntry
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

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}

		var page struct {
			Values []diffstatRaw `json:"values"`
			Next   string        `json:"next"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		resp.Body.Close()

		for _, v := range page.Values {
			path := v.New.Path
			if path == "" {
				path = v.Old.Path
			}
			entries = append(entries, diffstatEntry{
				Status: v.Status,
				Path:   path,
			})
		}

		url = page.Next
	}

	return entries, nil
}

func splitDiff(fullDiff string) map[string]string {
	patches := make(map[string]string)
	chunks := strings.Split(fullDiff, "diff --git ")

	for _, chunk := range chunks[1:] {
		newlineIdx := strings.Index(chunk, "\n")
		if newlineIdx == -1 {
			continue
		}
		header := chunk[:newlineIdx]
		parts := strings.Fields(header)
		if len(parts) < 2 {
			continue
		}
		path := strings.TrimPrefix(parts[1], "b/")
		patches[path] = "diff --git " + chunk
	}

	return patches
}

func bitbucketAPIError(resp *http.Response) error {
	var apiError struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&apiError); err == nil && apiError.Error.Message != "" {
		return fmt.Errorf("bitbucket API returned %s: %s", resp.Status, apiError.Error.Message)
	}

	return fmt.Errorf("bitbucket API returned %s", resp.Status)
}
