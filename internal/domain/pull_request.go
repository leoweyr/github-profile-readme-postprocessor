package domain

import (
	"time"
)

// PullRequest represents a proposed change to a repository.
type PullRequest struct {
	Number         int
	Title          string
	RepositoryName string
	HTMLURL        string
	State          string
	CreatedAt      time.Time
	MergedAt       *time.Time
	ClosedAt       *time.Time
}
