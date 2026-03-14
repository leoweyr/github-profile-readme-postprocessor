package domain

import (
	"time"
)

// ContributedRepository represents a repository that the user has contributed to, with activity metadata.
type ContributedRepository struct {
	Repository     *Repository
	ActiveAt       time.Time
	IsOwner        bool
	ActivityStats  *ActivityStats
	LatestActivity *ActivityItem
}
