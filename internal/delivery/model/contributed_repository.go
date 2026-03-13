package model

import (
	"time"
)

// ContributedRepository represents a repository that the user has contributed to.
type ContributedRepository struct {
	// Name is the full name of the repository (owner/repo).
	Name string `json:"name"`

	// ActiveAt is the timestamp of the most recent activity in this repository.
	ActiveAt time.Time `json:"active_at"`

	// Description is the description of the repository.
	Description string `json:"description"`

	// IsOwner indicates whether the user is the owner of the repository.
	IsOwner bool `json:"is_owner"`
}
