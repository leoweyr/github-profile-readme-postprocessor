package domain

// ActivityStats represents the statistics of user contributions to a repository.
type ActivityStats struct {
	CommitCount      int
	IssueCount       int
	PullRequestCount int
	TimeWindow       int // The time window in hours that these stats represent (e.g., 24, 168, 720, 8760).
}
