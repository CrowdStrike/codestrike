package scm

import (
	"context"
	"time"
)

type PullRequestFile struct {
	Filename string
	Status   string
	Patch    string
	Content  string
}

type ReviewComment struct {
	File string
	Line int
	Body string
}

type PRComment struct {
	ID        int64
	Author    string
	Body      string
	CreatedAt time.Time
	Path      string
	Line      int
	InReplyTo int64
}

type PRMetadata struct {
	Title          string
	Body           string
	CommitMessages []string
	BaseBranch     string
	HeadBranch     string
}

type Client interface {
	GetPullRequestDiff(ctx context.Context, number int) (string, error)
	GetPullRequestFiles(ctx context.Context, number int) ([]PullRequestFile, error)
	GetPullRequestMetadata(ctx context.Context, number int) (PRMetadata, error)
	GetFileContent(ctx context.Context, path, ref string) (string, error)
	PullRequestExists(ctx context.Context, number int) (bool, error)
	PublishComment(ctx context.Context, number int, body string) error
	GetPRComments(ctx context.Context, number int) ([]PRComment, error)
	GetPRReviewComments(ctx context.Context, number int) ([]PRComment, error)
}
