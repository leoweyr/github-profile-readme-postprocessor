package domain

import (
	"time"
)

// Repository represents a GitHub repository with relevant metadata.
type Repository struct {
	Name        string
	FullName    string
	HTMLURL     string
	StarredAt   time.Time
	CommitCount int
	Topics      []string
}
