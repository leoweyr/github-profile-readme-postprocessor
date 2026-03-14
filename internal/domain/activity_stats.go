package domain

// ActivityStats represents the statistics of user contributions to a repository.
type ActivityStats struct {
	CommitCount      int
	IssueCount       int
	PullRequestCount int
}
