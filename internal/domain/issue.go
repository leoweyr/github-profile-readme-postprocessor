package domain

import (
	"time"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain/enums"
)

// Issue represents a GitHub issue event (creation or comment) by a user.
type Issue struct {
	Title          string
	HTMLURL        string
	RepositoryName string
	CreatedAt      time.Time
	Number         int
	Action         enums.IssueAction
}
