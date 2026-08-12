package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/CrowdStrike/codestrike/internal/scm"
)

var _ scm.Client = (*Client)(nil)

type Config struct {
	Owner   string
	Repo    string
	Token   string
	BaseURL string
}

type Client struct {
	config     Config
	httpClient *http.Client
}

func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.github.com"
	}
	return &Client{
		config:     cfg,
		httpClient: http.DefaultClient,
	}
}

func NewWithHTTPClient(cfg Config, httpClient *http.Client) *Client {
	c := New(cfg)
	c.httpClient = httpClient
	return c
}

func (c *Client) GetPullRequestDiff(ctx context.Context, number int) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.config.BaseURL, c.config.Owner, c.config.Repo, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3.diff")
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

func (c *Client) PullRequestExists(ctx context.Context, number int) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.config.BaseURL, c.config.Owner, c.config.Repo, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
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

func (c *Client) PublishComment(ctx context.Context, number int, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.config.BaseURL, c.config.Owner, c.config.Repo, number)

	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return fmt.Errorf("marshaling comment: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
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
		return githubAPIError(resp)
	}

	return nil
}

func githubAPIError(resp *http.Response) error {
	var apiError struct {
		Message          string `json:"message"`
		DocumentationURL string `json:"documentation_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&apiError); err == nil && apiError.Message != "" {
		if apiError.DocumentationURL != "" {
			return fmt.Errorf("GitHub API returned %s: %s (%s)", resp.Status, apiError.Message, apiError.DocumentationURL)
		}
		return fmt.Errorf("GitHub API returned %s: %s", resp.Status, apiError.Message)
	}

	return fmt.Errorf("GitHub API returned %s", resp.Status)
}

func (c *Client) GetPullRequestFiles(ctx context.Context, number int) ([]scm.PullRequestFile, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/files", c.config.BaseURL, c.config.Owner, c.config.Repo, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("pull request #%d not found", number)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var ghFiles []struct {
		Filename string `json:"filename"`
		Status   string `json:"status"`
		Patch    string `json:"patch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghFiles); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	files := make([]scm.PullRequestFile, 0, len(ghFiles))
	for _, f := range ghFiles {
		files = append(files, scm.PullRequestFile{
			Filename: f.Filename,
			Status:   f.Status,
			Patch:    f.Patch,
		})
	}

	return files, nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}
}
