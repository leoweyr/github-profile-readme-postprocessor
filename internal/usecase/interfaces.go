package usecase

import (
	"context"
	"time"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// CommitFetcher defines the interface for fetching user commits.
type CommitFetcher interface {
	FetchCommits(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Commit, error)
	FetchPrivateCommits(ctx context.Context, username string, privateRepos []*domain.Repository, startTime, endTime time.Time) ([]*domain.Commit, error)
}

// IssueFetcher defines the interface for fetching user issues.
type IssueFetcher interface {
	FetchIssueActivities(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Issue, error)
	FetchPrivateIssueActivities(ctx context.Context, username string, privateRepos []*domain.Repository, startTime, endTime time.Time) ([]*domain.Issue, error)
}

// PullRequestFetcher defines the interface for fetching user pull requests.
type PullRequestFetcher interface {
	FetchPullRequests(context context.Context, username string, startTime, endTime time.Time) ([]*domain.PullRequest, error)
	FetchPrivatePullRequests(ctx context.Context, username string, privateRepos []*domain.Repository, startTime, endTime time.Time) ([]*domain.PullRequest, error)
}

// RepositoryFetcher defines the interface for fetching repository details.
type RepositoryFetcher interface {
	FetchRepository(context context.Context, owner, repositoryName string) (*domain.Repository, error)
	FetchPrivateRepositories(ctx context.Context) ([]*domain.Repository, error)
}
