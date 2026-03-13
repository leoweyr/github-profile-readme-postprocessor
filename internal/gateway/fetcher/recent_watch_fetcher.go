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

// RecentWatchFetcher handles the retrieval of recent watches for a user.
type RecentWatchFetcher struct {
	client *github.Client
}

// NewRecentWatchFetcher creates a new instance of RecentWatchFetcher with the provided token.
func NewRecentWatchFetcher(token string) *RecentWatchFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	var httpClient *http.Client = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &RecentWatchFetcher{
		client: client,
	}
}

// FetchRecentWatches retrieves recent watches for the user within the specified time range.
func (fetcher *RecentWatchFetcher) FetchRecentWatches(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Repository, error) {
	var allWatchedRepositories []*domain.Repository
	var listOptions *github.ListOptions = &github.ListOptions{PerPage: 100}

	for {
		var events []*github.Event
		var response *github.Response
		var listError error

		events, response, listError = fetcher.client.Activity.ListEventsPerformedByUser(context, username, false, listOptions)

		if listError != nil {
			return nil, fmt.Errorf("failed to list user events: %w", listError)
		}

		var event *github.Event

		for _, event = range events {
			var createdAt time.Time = event.GetCreatedAt().Time

			if createdAt.After(endTime) {
				continue
			}

			if createdAt.Before(startTime) {
				return allWatchedRepositories, nil
			}

			if event.GetType() == "WatchEvent" {
				var repositoryName string = event.GetRepo().GetName()
				var watchedAt time.Time = event.GetCreatedAt().Time

				var watchedRepository *domain.Repository = &domain.Repository{
					Name:      repositoryName,
					FullName:  repositoryName,
					HTMLURL:   fmt.Sprintf("https://github.com/%s", repositoryName),
					StarredAt: watchedAt,
					Topics:    []string{},
				}

				allWatchedRepositories = append(allWatchedRepositories, watchedRepository)
			}
		}

		if response.NextPage == 0 {
			break
		}

		listOptions.Page = response.NextPage
	}

	return allWatchedRepositories, nil
}
