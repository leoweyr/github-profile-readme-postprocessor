package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// PushFetcher handles the retrieval of repositories pushed to by a user.
type PushFetcher struct {
	client *github.Client
}

// NewPushFetcher creates a new instance of PushFetcher with the provided token.
func NewPushFetcher(token string) *PushFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	var httpClient = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &PushFetcher{
		client: client,
	}
}

// FetchPushes retrieves repositories pushed to by the user within the specified time range.
func (fetcher *PushFetcher) FetchPushes(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Repository, error) {
	var allPushedRepositories []*domain.Repository
	var listOptions *github.RepositoryListOptions = &github.RepositoryListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		var repositories []*github.Repository
		var response *github.Response
		var err error

		// List repositories for the specified user.
		repositories, response, err = fetcher.client.Repositories.List(context, username, listOptions)

		if err != nil {
			return nil, fmt.Errorf("failed to list user repositories: %w", err)
		}

		for _, repository := range repositories {
			if repository.PushedAt == nil {
				continue
			}

			var pushedAt time.Time = repository.PushedAt.Time

			// Skip the repository if the latest push is after the end time.
			if pushedAt.After(endTime) {
				continue
			}

			// Stop fetching if the repository was last pushed before the start time.
			if pushedAt.Before(startTime) {
				return allPushedRepositories, nil
			}

			// Get commit count for this repository within the range.
			var commitCount int
			var countErr error
			commitCount, countErr = fetcher.countCommits(context, username, repository.GetName(), startTime, endTime)

			if countErr != nil {
				// Log error but continue.
				fmt.Printf("Error counting commits for %s: %v\n", repository.GetName(), countErr)
				continue
			}

			if commitCount == 0 {
				continue
			}

			var repositoryName string = repository.GetName()
			var repositoryFullName string = repository.GetFullName()
			var repositoryHTMLURL string = repository.GetHTMLURL()

			var domainRepository *domain.Repository = &domain.Repository{
				Name:        repositoryName,
				FullName:    repositoryFullName,
				HTMLURL:     repositoryHTMLURL,
				StarredAt:   pushedAt, // Use the latest push time as the timestamp.
				CommitCount: commitCount,
				Topics:      repository.Topics,
			}

			allPushedRepositories = append(allPushedRepositories, domainRepository)
		}

		if response.NextPage == 0 {
			break
		}

		listOptions.Page = response.NextPage
	}

	return allPushedRepositories, nil
}

// countCommits returns the number of commits by the user in the repository within the time range.
func (fetcher *PushFetcher) countCommits(context context.Context, username, repositoryName string, startTime, endTime time.Time) (int, error) {
	var commitsListOptions *github.CommitsListOptions = &github.CommitsListOptions{
		Author: username,
		Since:  startTime,
		Until:  endTime,
		ListOptions: github.ListOptions{
			PerPage: 1, // Only need 1 item to get the total count from pagination.
		},
	}

	var commits []*github.RepositoryCommit
	var response *github.Response
	var err error

	commits, response, err = fetcher.client.Repositories.ListCommits(context, username, repositoryName, commitsListOptions)
	if err != nil {
		return 0, err
	}

	// If no commits are found, return 0.
	if len(commits) == 0 {
		return 0, nil
	}

	// If there is only one page, the length of commits is the total count.
	if response.LastPage == 0 {
		return len(commits), nil
	}

	// Otherwise, LastPage indicates the total number of pages (commits), since PerPage is 1.
	return response.LastPage, nil
}
