package version

import "testing"

func TestShort(t *testing.T) {
	origVersion, origCommit, origCommitDate := Version, Commit, CommitDate
	defer func() { Version, Commit, CommitDate = origVersion, origCommit, origCommitDate }()

	Version = "1.2.3"
	Commit = "abcdef1234567890"
	CommitDate = "2026-08-11"

	want := "1.2.3 (abcdef1 2026-08-11)"
	if got := Short(); got != want {
		t.Errorf("Short() = %q, want %q", got, want)
	}
}

func TestShort_ShortCommit(t *testing.T) {
	origVersion, origCommit, origCommitDate := Version, Commit, CommitDate
	defer func() { Version, Commit, CommitDate = origVersion, origCommit, origCommitDate }()

	Version = "dev"
	Commit = "abc123"
	CommitDate = "unknown"

	want := "dev (abc123 unknown)"
	if got := Short(); got != want {
		t.Errorf("Short() = %q, want %q", got, want)
	}
}
