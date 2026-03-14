package domain

import (
	"time"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain/enums"
)

// ActivityType represents the type of activity.
type ActivityType string

const (
	// ActivityTypeCommit represents a commit activity.
	ActivityTypeCommit ActivityType = "commit"
	// ActivityTypeIssue represents an issue activity.
	ActivityTypeIssue ActivityType = "issue"
	// ActivityTypePullRequest represents a pull request activity.
	ActivityTypePullRequest ActivityType = "pull_request"
)

// ActivityItem represents a single activity item (e.g., a commit, issue, or pull request).
type ActivityItem struct {
	Type        ActivityType
	Title       string // Message for commit, Title for Issue/PR.
	URL         string
	CreatedAt   time.Time
	IssueAction enums.IssueAction // Only for Issue.
}
