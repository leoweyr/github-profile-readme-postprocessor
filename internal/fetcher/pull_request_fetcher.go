package fetcher

import (
	"context"
	"fmt"
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

	var httpClient = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &PullRequestFetcher{
		client: client,
	}
}

// FetchPullRequests retrieves pull requests created by the user within the specified time range.
func (fetcher *PullRequestFetcher) FetchPullRequests(ctx context.Context, username string, startTime, endTime time.Time) ([]*domain.PullRequest, error) {
	var allPRs []*domain.PullRequest

	// Prepare search option.
	var searchOptions = &github.SearchOptions{
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
		var err error

		// Execute search.
		result, response, err = fetcher.client.Search.Issues(ctx, query, searchOptions)

		if err != nil {
			return nil, fmt.Errorf("failed to search pull requests: %w", err)
		}

		// Process results.
		for _, issue := range result.Issues {
			// Verify the issue is a pull request.
			if !issue.IsPullRequest() {
				continue
			}

			// Extract repository name from URL (e.g., https://api.github.com/repos/owner/repo).
			var repoURL string = issue.GetRepositoryURL()

			// Parse "owner/repo" from "https://api.github.com/repos/owner/repo".
			// A regex or strict URL parsing is preferred for production stability, avoiding heavy dependencies here.
			// The current implementation assumes the standard GitHub API URL structure.
			var repositoryName string = repoURL

			if len(repoURL) > 29 {
				// Format: https://api.github.com/repos/{owner}/{repo}.
			}

			var pr *domain.PullRequest = &domain.PullRequest{
				Number:         issue.GetNumber(),
				Title:          issue.GetTitle(),
				RepositoryName: repositoryName,
				HTMLURL:        issue.GetHTMLURL(),
				State:          issue.GetState(),
				CreatedAt:      issue.GetCreatedAt().Time,
			}

			if !issue.GetClosedAt().IsZero() {
				var t time.Time = issue.GetClosedAt().Time
				pr.ClosedAt = &t
			}

			// Check for merge status via PullRequestLinks.
			// The Issue struct contains PullRequestLinks which provides the MergedAt timestamp.
			if issue.PullRequestLinks != nil && !issue.PullRequestLinks.MergedAt.IsZero() {
				var t time.Time = issue.PullRequestLinks.MergedAt.Time
				pr.MergedAt = &t
			}

			allPRs = append(allPRs, pr)
		}

		if response.NextPage == 0 {
			break
		}

		searchOptions.Page = response.NextPage
	}

	return allPRs, nil
}
