package enums

// IssueAction represents the type of activity performed on an issue.
type IssueAction string

const (
	// IssueActionCreated represents creating a new issue.
	IssueActionCreated IssueAction = "created"

	// IssueActionCommented represents commenting on an existing issue.
	IssueActionCommented IssueAction = "commented"
)
