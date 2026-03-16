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

// FetchPrivateCommits retrieves commits from specified private repositories for the user.
func (fetcher *CommitFetcher) FetchPrivateCommits(context context.Context, username string, privateRepos []*domain.Repository, startTime, endTime time.Time) ([]*domain.Commit, error) {
	var allPrivateCommits []*domain.Commit
	var repository *domain.Repository

	for _, repository = range privateRepos {
		// Optimization: Skip repositories that haven't been pushed to since the start time.
		if !repository.PushedAt.IsZero() && repository.PushedAt.Before(startTime) {
			continue
		}

		// Configure list options per repository.
		var currentRepoOptions *github.CommitsListOptions = &github.CommitsListOptions{
			Author: username, // Strict filtering by author as requested.
			Since:  startTime,
			Until:  endTime,
			ListOptions: github.ListOptions{
				PerPage: 100, // Increase page size for efficiency.
			},
		}

		// Parse owner/repo.
		var parts []string = strings.Split(repository.FullName, "/")

		if len(parts) != 2 {
			continue
		}

		var owner string = parts[0]
		var repositoryName string = parts[1]

		var commits []*github.RepositoryCommit
		var response *github.Response
		var listError error

		for {
			commits, response, listError = fetcher.client.Repositories.ListCommits(context, owner, repositoryName, currentRepoOptions)

			if listError != nil {
				// Log error but continue to next repo.
				fmt.Printf("DEBUG: Failed to list commits for private repo %s: %v\n", repository.FullName, listError)

				break
			}

			var item *github.RepositoryCommit

			for _, item = range commits {
				if item.Commit == nil {
					continue
				}

				var committedAt time.Time

				if item.Commit.Committer.Date != nil {
					committedAt = item.Commit.Committer.Date.Time
				}

				var commit *domain.Commit = &domain.Commit{
					SHA:            item.GetSHA(),
					Message:        item.Commit.GetMessage(),
					RepositoryName: repository.FullName,
					HTMLURL:        item.GetHTMLURL(),
					CommittedAt:    committedAt,
				}

				allPrivateCommits = append(allPrivateCommits, commit)
			}

			if response.NextPage == 0 {
				break
			}

			currentRepoOptions.ListOptions.Page = response.NextPage
		}
	}

	return allPrivateCommits, nil
}

// GetLatestPrivateCommit fetches the single most recent commit for a repo to determine precise activity time.
func (fetcher *CommitFetcher) GetLatestPrivateCommit(ctx context.Context, username string, repo *domain.Repository) (*domain.Commit, error) {
	// Parse owner/repo.
	var parts []string = strings.Split(repo.FullName, "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo name: %s", repo.FullName)
	}

	var owner string = parts[0]
	var repositoryName string = parts[1]

	var listOptions *github.CommitsListOptions = &github.CommitsListOptions{
		Author: username,
		ListOptions: github.ListOptions{
			PerPage: 1, // Only need the latest one.
			Page:    1,
		},
	}

	var commits []*github.RepositoryCommit
	var listError error
	commits, _, listError = fetcher.client.Repositories.ListCommits(ctx, owner, repositoryName, listOptions)

	if listError != nil {
		return nil, listError
	}

	if len(commits) == 0 {
		return nil, nil // No commits found.
	}

	var item *github.RepositoryCommit = commits[0]

	if item.Commit == nil {
		return nil, nil
	}

	var committedAt time.Time

	if item.Commit.Committer.Date != nil {
		committedAt = item.Commit.Committer.Date.Time
	}

	return &domain.Commit{
		SHA:            item.GetSHA(),
		Message:        item.Commit.GetMessage(),
		RepositoryName: repo.FullName,
		HTMLURL:        item.GetHTMLURL(),
		CommittedAt:    committedAt,
	}, nil
}

// CountPrivateCommits efficiently counts commits in a private repo within a time window using pagination metadata.
func (fetcher *CommitFetcher) CountPrivateCommits(ctx context.Context, username string, repo *domain.Repository, since, until time.Time) (int, error) {
	var parts []string = strings.Split(repo.FullName, "/")

	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid repo name: %s", repo.FullName)
	}

	var owner string = parts[0]
	var repositoryName string = parts[1]

	var listOptions *github.CommitsListOptions = &github.CommitsListOptions{
		Author: username,
		Since:  since,
		Until:  until,
		ListOptions: github.ListOptions{
			PerPage: 1, // Minimal payload to get the LastPage header.
			Page:    1,
		},
	}

	var commits []*github.RepositoryCommit
	var response *github.Response
	var listError error
	commits, response, listError = fetcher.client.Repositories.ListCommits(ctx, owner, repositoryName, listOptions)

	if listError != nil {
		return 0, listError
	}

	if len(commits) == 0 {
		return 0, nil
	}

	// If LastPage is 0, it means there is only 1 page.
	// Since PerPage is 1, the total count is the number of items returned (which is 1).
	// However, if there are multiple pages, LastPage tells us the total count because PerPage is 1.
	if response.LastPage == 0 {
		return len(commits), nil
	}

	return response.LastPage, nil
}
