package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// CommitFetcher handles the retrieval of commits made by a user.
type CommitFetcher struct {
	client *github.Client
}

// NewCommitFetcher creates a new instance of CommitFetcher with the provided token.
func NewCommitFetcher(token string) *CommitFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	var httpClient *http.Client = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &CommitFetcher{
		client: client,
	}
}

// FetchCommits retrieves commits authored by the user within the specified time range.
func (fetcher *CommitFetcher) FetchCommits(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Commit, error) {
	var allCommits []*domain.Commit
	var searchOptions *github.SearchOptions = &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "committer-date",
		Order:       "desc",
	}

	// Format the date range for the search query (YYYY-MM-DD).
	var dateRange string = fmt.Sprintf("%s..%s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	var query string = fmt.Sprintf("author:%s committer-date:%s", username, dateRange)

	for {
		var result *github.CommitsSearchResult
		var response *github.Response
		var searchError error

		// Search for commits matching the query.
		// Note: The Search API has rate limits (30 requests per minute for authenticated users).
		result, response, searchError = fetcher.client.Search.Commits(context, query, searchOptions)

		if searchError != nil {
			return nil, fmt.Errorf("failed to search commits: %w", searchError)
		}

		var item *github.CommitResult
		for _, item = range result.Commits {
			if item.Commit == nil || item.Commit.Committer == nil {
				continue
			}

			var commitHash string = item.GetSHA()
			var message string = item.Commit.GetMessage()
			var htmlURL string = item.GetHTMLURL()
			var committedAt time.Time

			if item.Commit.Committer.Date != nil {
				committedAt = item.Commit.Committer.Date.Time
			}

			var repositoryName string = ""

			if item.Repository != nil {
				repositoryName = item.Repository.GetFullName()
			}

			var commit *domain.Commit = &domain.Commit{
				SHA:            commitHash,
				Message:        message,
				RepositoryName: repositoryName,
				HTMLURL:        htmlURL,
				CommittedAt:    committedAt,
			}

			allCommits = append(allCommits, commit)
		}

		if response.NextPage == 0 {
			break
		}

		searchOptions.Page = response.NextPage
	}

	return allCommits, nil
}
