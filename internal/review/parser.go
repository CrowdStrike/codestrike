package review

import (
	"fmt"
	"strconv"
	"strings"
)

type PRReference struct {
	Owner  string
	Repo   string
	Number int
}

func ParsePRReference(owner, repo string, number int) PRReference {
	return PRReference{Owner: owner, Repo: repo, Number: number}
}

func ParsePRURL(url string) (PRReference, error) {
	url = strings.TrimRight(url, "/")

	// Expected: https://github.com/{owner}/{repo}/pull/{number}
	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return PRReference{}, fmt.Errorf("invalid PR URL: %s", url)
	}

	pullIdx := -1
	for i, p := range parts {
		if p == "pull" {
			pullIdx = i
			break
		}
	}
	if pullIdx == -1 || pullIdx < 2 {
		return PRReference{}, fmt.Errorf("invalid PR URL: missing /pull/ segment")
	}

	owner := parts[pullIdx-2]
	repo := parts[pullIdx-1]
	numberStr := parts[pullIdx+1]

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return PRReference{}, fmt.Errorf("invalid PR number %q: %w", numberStr, err)
	}

	return PRReference{Owner: owner, Repo: repo, Number: number}, nil
}
