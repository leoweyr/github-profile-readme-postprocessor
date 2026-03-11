package domain

import (
	"time"
)

// Repository represents a GitHub repository with relevant metadata.
type Repository struct {
	Name        string
	FullName    string
	Description string
	HTMLURL     string
	Owner       string
	Topics      []string
	StarredAt   time.Time
}
