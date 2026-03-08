package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// RecentWatchFetcher handles the retrieval of recently watched repositories for a user.
// Note: This fetcher relies on the GitHub Events API, which only provides access to events
// from the last 90 days or the last 300 events. Older watch events cannot be retrieved.
type RecentWatchFetcher struct {
	client *github.Client
}

// NewRecentWatchFetcher creates a new instance of RecentWatchFetcher with the provided token.
func NewRecentWatchFetcher(token string) *RecentWatchFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	var httpClient = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &RecentWatchFetcher{
		client: client,
	}
}

// FetchRecentWatches retrieves repositories watched by the user within the specified time range.
// The time range is effectively limited by GitHub's Events API retention (90 days / 300 events).
func (fetcher *RecentWatchFetcher) FetchRecentWatches(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Repository, error) {
	var allWatchedRepositories []*domain.Repository
	var listOptions *github.ListOptions = &github.ListOptions{PerPage: 100}

	for {
		var events []*github.Event
		var response *github.Response
		var err error

		// List events performed by a user.
		events, response, err = fetcher.client.Activity.ListEventsPerformedByUser(context, username, false, listOptions)

		if err != nil {
			return nil, fmt.Errorf("failed to list user events: %w", err)
		}

		for _, event := range events {
			var createdAt time.Time = event.GetCreatedAt().Time

			// Skip the event if it is newer than the end time.
			if createdAt.After(endTime) {
				continue
			}

			// Stop fetching if the event is older than the start time.
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
