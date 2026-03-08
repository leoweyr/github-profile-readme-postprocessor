package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// ForkFetcher handles the retrieval of forked repositories for a user.
type ForkFetcher struct {
	client *github.Client
}

// NewForkFetcher creates a new instance of ForkFetcher with the provided token.
func NewForkFetcher(token string) *ForkFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	var httpClient = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &ForkFetcher{
		client: client,
	}
}

// FetchForks retrieves repositories forked by the user within the specified time range.
func (fetcher *ForkFetcher) FetchForks(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Repository, error) {
	var allForkedRepositories []*domain.Repository
	var listOptions *github.RepositoryListOptions = &github.RepositoryListOptions{
		Sort:        "created",
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
			var createdAt time.Time = repository.GetCreatedAt().Time

			// Skip the repository if it is newer than the end time.
			if createdAt.After(endTime) {
				continue
			}

			// Stop fetching if the repository is older than the start time.
			if createdAt.Before(startTime) {
				return allForkedRepositories, nil
			}

			// Check if the repository is a fork.
			if repository.GetFork() {
				var repositoryName string = repository.GetName()
				var repositoryFullName string = repository.GetFullName()
				var repositoryHTMLURL string = repository.GetHTMLURL()

				var domainRepository *domain.Repository = &domain.Repository{
					Name:      repositoryName,
					FullName:  repositoryFullName,
					HTMLURL:   repositoryHTMLURL,
					StarredAt: createdAt, // For forks, CreatedAt is when the fork was made.
				}

				allForkedRepositories = append(allForkedRepositories, domainRepository)
			}
		}

		if response.NextPage == 0 {
			break
		}

		listOptions.Page = response.NextPage
	}

	return allForkedRepositories, nil
}
