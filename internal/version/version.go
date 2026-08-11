package version

import "fmt"

// Short returns the compact version string in the form "x.y.z (abc1234 2006-01-02)".
func Short() string {
	commit := Commit
	if len(commit) > 7 {
		commit = commit[:7]
	}
	return fmt.Sprintf("%s (%s %s)", Version, commit, CommitDate)
}

var (
	// Version specifies the version of the software at the time of compilation.
	// This value is automatically populated by GoReleaser, reflecting the version tag
	// associated with the source code used for the build.
	Version = "dev"

	// Commit represents the specific git commit hash of the software at the time of compilation.
	// This is set by GoReleaser and helps in identifying the exact state of the source code
	// used for building the software, facilitating better traceability and debugging.
	Commit = "unknown"

	// CommitDate represents the specific git commit date of the software at the time of compilation.
	// This is set by GoReleaser and helps in identifying the exact state of the source code
	// used for building the software, facilitating better traceability and debugging.
	CommitDate = "unknown"

	// BuiltBy indicates the username or identifier of the person or system that compiled the software.
	// This information is filled out by GoReleaser and can be useful in understanding
	// or tracking who performed the build, especially in environments with multiple developers or CI systems.
	BuiltBy = "unknown"

	// Date captures the date and time when the software was compiled.
	// This is automatically populated by GoReleaser. Knowing the build date is useful for
	// contextualizing the software version and assessing its recency, especially when diagnosing issues
	// or managing updates.
	Date = "unknown"
)
