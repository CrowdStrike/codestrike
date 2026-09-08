package review

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ProviderGitHub    = "github"
	ProviderBitbucket = "bitbucket"
)

type PRReference struct {
	Provider string
	Owner    string
	Repo     string
	Number   int
}

func ParsePrReference(owner, repo string, number int) PRReference {
	return PRReference{
		Provider: ProviderGitHub, 
		Owner: owner, 
		Repo: repo, 
		Number: number,
	}
}

func ParsePrUrl(url string) (PRReference, error) {
	url = strings.TrimRight(url, "/")
	parts := strings.Split(url, "/")

	if idx := findSegment(parts, "pull-requests"); idx != -1 {
		return parsePrParts(parts, idx, ProviderBitbucket)
	}
	if idx := findSegment(parts, "pull"); idx != -1 {
		return parsePrParts(parts, idx, ProviderGitHub)
	}

	return PRReference{}, fmt.Errorf("invalid PR URL: missing /pull/ or /pull-requests/ segment")
}

func parsePrParts(parts []string, pullIdx int, provider string) (PRReference, error) {
	if pullIdx < 2 || pullIdx+1 >= len(parts) {
		return PRReference{}, fmt.Errorf("invalid PR URL: not enough segments")
	}

	owner := parts[pullIdx-2]
	repo := parts[pullIdx-1]
	if owner == "" || repo == "" {
		return PRReference{}, fmt.Errorf("invalid PR URL: missing owner or repo")
	}

	numberStr := parts[pullIdx+1]

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return PRReference{}, fmt.Errorf("invalid PR number %q: %w", numberStr, err)
	}

	return PRReference{Provider: provider, Owner: owner, Repo: repo, Number: number}, nil
}

func findSegment(parts []string, segment string) int {
	for i, p := range parts {
		if p == segment {
			return i
		}
	}
	return -1
}
