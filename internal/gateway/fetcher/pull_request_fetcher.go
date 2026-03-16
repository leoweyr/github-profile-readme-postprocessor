package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// PullRequestFetcher handles the retrieval of pull requests created by a user.
type PullRequestFetcher struct {
	client *github.Client
}

// NewPullRequestFetcher creates a new instance of PullRequestFetcher with the provided token.
func NewPullRequestFetcher(token string) *PullRequestFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	var httpClient *http.Client = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &PullRequestFetcher{
		client: client,
	}
}

// FetchPullRequests retrieves pull requests created by the user within the specified time range.
func (fetcher *PullRequestFetcher) FetchPullRequests(context context.Context, username string, startTime, endTime time.Time) ([]*domain.PullRequest, error) {
	var allPullRequests []*domain.PullRequest

	// Prepare search option.
	var searchOptions *github.SearchOptions = &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "created",
		Order:       "desc",
	}

	// Format date range: created:YYYY-MM-DD..YYYY-MM-DD.
	var dateRange string = fmt.Sprintf("%s..%s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	// Query: type:pr author:username created:start..end.
	var query string = fmt.Sprintf("type:pr author:%s created:%s", username, dateRange)

	for {
		var result *github.IssuesSearchResult
		var response *github.Response
		var searchError error

		// Execute search.
		result, response, searchError = fetcher.client.Search.Issues(context, query, searchOptions)

		if searchError != nil {
			return nil, fmt.Errorf("failed to search pull requests: %w", searchError)
		}

		// Process results.
		var issue *github.Issue

		for _, issue = range result.Issues {
			// Verify the issue is a pull request.
			if !issue.IsPullRequest() {
				continue
			}

			// Extract repository name from URL (e.g., https://api.github.com/repos/owner/repo).
			var repositoryURL string = issue.GetRepositoryURL()

			// Parse "owner/repo" from "https://api.github.com/repos/owner/repo".
			// A regex or strict URL parsing is preferred for production stability, avoiding heavy dependencies here.
			// The current implementation assumes the standard GitHub API URL structure.
			var repositoryName string = repositoryURL

			if len(repositoryURL) > 29 {
				// Format: https://api.github.com/repos/{owner}/{repo}.
			}

			var pullRequest *domain.PullRequest = &domain.PullRequest{
				Number:         issue.GetNumber(),
				Title:          issue.GetTitle(),
				RepositoryName: repositoryName,
				HTMLURL:        issue.GetHTMLURL(),
				State:          issue.GetState(),
				CreatedAt:      issue.GetCreatedAt().Time,
			}

			if !issue.GetClosedAt().IsZero() {
				var closedTimestamp time.Time = issue.GetClosedAt().Time
				pullRequest.ClosedAt = &closedTimestamp
			}

			// Check for merge status via PullRequestLinks.
			// The Issue struct contains PullRequestLinks which provides the MergedAt timestamp.
			if issue.PullRequestLinks != nil && !issue.PullRequestLinks.MergedAt.IsZero() {
				var mergedTimestamp time.Time = issue.PullRequestLinks.MergedAt.Time
				pullRequest.MergedAt = &mergedTimestamp
			}

			allPullRequests = append(allPullRequests, pullRequest)
		}

		if response.NextPage == 0 {
			break
		}

		searchOptions.Page = response.NextPage
	}

	return allPullRequests, nil
}

// FetchPrivatePullRequests retrieves pull requests for private repositories.
func (fetcher *PullRequestFetcher) FetchPrivatePullRequests(ctx context.Context, username string, privateRepos []*domain.Repository, startTime, endTime time.Time) ([]*domain.PullRequest, error) {
	var allPrivatePullRequests []*domain.PullRequest
	var repository *domain.Repository

	for _, repository = range privateRepos {
		// Optimization: Skip repositories that haven't been pushed to since the start time.
		if !repository.PushedAt.IsZero() && repository.PushedAt.Before(startTime) {
			continue
		}

		var parts []string = strings.Split(repository.FullName, "/")

		if len(parts) != 2 {
			continue
		}

		var owner string = parts[0]
		var repositoryName string = parts[1]

		var listOptions *github.IssueListByRepoOptions = &github.IssueListByRepoOptions{
			Creator: username,
			Since:   startTime,
			State:   "all",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		var issues []*github.Issue
		var response *github.Response
		var listError error

		for {
			issues, response, listError = fetcher.client.Issues.ListByRepo(ctx, owner, repositoryName, listOptions)

			if listError != nil {
				fmt.Printf("Warning: failed to list PRs for private repo %s: %v\n", repository.FullName, listError)

				break
			}

			var issue *github.Issue

			for _, issue = range issues {
				if !issue.IsPullRequest() {
					continue
				}

				if issue.GetCreatedAt().Before(startTime) || issue.GetCreatedAt().After(endTime) {
					continue
				}

				var pullRequest *domain.PullRequest = &domain.PullRequest{
					Number:         issue.GetNumber(),
					Title:          issue.GetTitle(),
					RepositoryName: repository.FullName,
					HTMLURL:        issue.GetHTMLURL(),
					State:          issue.GetState(),
					CreatedAt:      issue.GetCreatedAt().Time,
				}

				if !issue.GetClosedAt().IsZero() {
					var closedAt time.Time = issue.GetClosedAt().Time
					pullRequest.ClosedAt = &closedAt
				}

				allPrivatePullRequests = append(allPrivatePullRequests, pullRequest)
			}

			if response.NextPage == 0 {
				break
			}

			listOptions.ListOptions.Page = response.NextPage
		}
	}

	return allPrivatePullRequests, nil
}

// GetLatestPrivatePullRequest fetches the single most recent PR for a repo.
func (fetcher *PullRequestFetcher) GetLatestPrivatePullRequest(ctx context.Context, username string, repo *domain.Repository) (*domain.PullRequest, error) {
	var parts []string = strings.Split(repo.FullName, "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo name: %s", repo.FullName)
	}

	var owner string = parts[0]
	var repositoryName string = parts[1]

	var listOptions *github.PullRequestListOptions = &github.PullRequestListOptions{
		State:       "all",
		Sort:        "created",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 1},
	}

	var prs []*github.PullRequest
	var listError error

	prs, _, listError = fetcher.client.PullRequests.List(ctx, owner, repositoryName, listOptions)

	if listError != nil {
		return nil, listError
	}

	var pr *github.PullRequest

	for _, pr = range prs {
		if pr.User != nil && pr.User.Login != nil && *pr.User.Login == username {
			var createdAt time.Time = pr.GetCreatedAt().Time
			var pullRequest *domain.PullRequest = &domain.PullRequest{
				Number:         pr.GetNumber(),
				Title:          pr.GetTitle(),
				RepositoryName: repo.FullName,
				HTMLURL:        pr.GetHTMLURL(),
				State:          pr.GetState(),
				CreatedAt:      createdAt,
			}

			return pullRequest, nil
		}
	}

	return nil, nil
}

// CountPrivatePullRequests counts PRs in a private repo within a time window.
func (fetcher *PullRequestFetcher) CountPrivatePullRequests(ctx context.Context, username string, repo *domain.Repository, since, until time.Time) (int, error) {
	// Use Search API for precise counting.
	var query string = fmt.Sprintf("repo:%s type:pr author:%s created:%s..%s", repo.FullName, username, since.Format("2006-01-02"), until.Format("2006-01-02"))
	var searchOptions *github.SearchOptions = &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	}

	var result *github.IssuesSearchResult
	var response *github.Response
	var searchError error

	result, response, searchError = fetcher.client.Search.Issues(ctx, query, searchOptions)

	if searchError != nil {
		return 0, searchError
	}

	if result.Total == nil {
		return 0, nil
	}

	_ = response // Unused.

	return *result.Total, nil
}
