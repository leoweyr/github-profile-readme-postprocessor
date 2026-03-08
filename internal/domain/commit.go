package domain

import (
	"time"
)

// Commit represents a specific code change in a repository.
type Commit struct {
	SHA            string
	Message        string
	RepositoryName string
	HTMLURL        string
	CommittedAt    time.Time
}
