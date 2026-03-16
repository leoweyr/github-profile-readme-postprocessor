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
	GetLatestPrivateCommit(ctx context.Context, username string, repo *domain.Repository) (*domain.Commit, error)
	CountPrivateCommits(ctx context.Context, username string, repo *domain.Repository, since, until time.Time) (int, error)
}

// IssueFetcher defines the interface for fetching user issues.
type IssueFetcher interface {
	FetchIssueActivities(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Issue, error)
	FetchPrivateIssueActivities(ctx context.Context, username string, privateRepos []*domain.Repository, startTime, endTime time.Time) ([]*domain.Issue, error)
	GetLatestPrivateIssue(ctx context.Context, username string, repo *domain.Repository) (*domain.Issue, error)
	CountPrivateIssues(ctx context.Context, username string, repo *domain.Repository, since, until time.Time) (int, error)
}

// PullRequestFetcher defines the interface for fetching user pull requests.
type PullRequestFetcher interface {
	FetchPullRequests(context context.Context, username string, startTime, endTime time.Time) ([]*domain.PullRequest, error)
	FetchPrivatePullRequests(ctx context.Context, username string, privateRepos []*domain.Repository, startTime, endTime time.Time) ([]*domain.PullRequest, error)
	GetLatestPrivatePullRequest(ctx context.Context, username string, repo *domain.Repository) (*domain.PullRequest, error)
	CountPrivatePullRequests(ctx context.Context, username string, repo *domain.Repository, since, until time.Time) (int, error)
}

// RepositoryFetcher defines the interface for fetching repository details.
type RepositoryFetcher interface {
	FetchRepository(context context.Context, owner, repositoryName string) (*domain.Repository, error)
	FetchPrivateRepositories(ctx context.Context) ([]*domain.Repository, error)
}
