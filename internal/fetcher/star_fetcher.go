package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// StarFetcher handles the retrieval of starred repositories for a user.
type StarFetcher struct {
	client *githubv4.Client
}

// NewStarFetcher creates a new instance of StarFetcher with the provided token.
func NewStarFetcher(token string) *StarFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	var httpClient = oauth2.NewClient(context.Background(), tokenSource)
	var client *githubv4.Client = githubv4.NewClient(httpClient)

	return &StarFetcher{
		client: client,
	}
}

// FetchStars retrieves repositories starred by the user within the specified time range.
func (fetcher *StarFetcher) FetchStars(context context.Context, startTime, endTime time.Time) ([]*domain.Repository, error) {
	var query struct {
		Viewer struct {
			StarredRepositories struct {
				Edges []struct {
					StarredAt time.Time
					Node      struct {
						Name          string
						NameWithOwner string
						Url           string
					}
				}
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"starredRepositories(first: 100, orderBy: {field: STARRED_AT, direction: DESC}, after: $cursor)"`
		}
	}

	var allStarredRepositories []*domain.Repository
	var variables map[string]interface{} = map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
	}

	for {
		var err error = fetcher.client.Query(context, &query, variables)

		if err != nil {
			return nil, fmt.Errorf("failed to query starred repositories: %w", err)
		}

		for _, edge := range query.Viewer.StarredRepositories.Edges {
			var starredAt time.Time = edge.StarredAt

			// If the star is newer than endTime, skip it but continue (since we sort desc, newer are first)
			if starredAt.After(endTime) {
				continue
			}

			// If the star is older than startTime, stop fetching (since we sort desc)
			if starredAt.Before(startTime) {
				return allStarredRepositories, nil
			}

			var starredRepository *domain.Repository = &domain.Repository{
				Name:      edge.Node.Name,
				FullName:  edge.Node.NameWithOwner,
				HTMLURL:   edge.Node.Url,
				StarredAt: starredAt,
			}

			allStarredRepositories = append(allStarredRepositories, starredRepository)
		}

		if !query.Viewer.StarredRepositories.PageInfo.HasNextPage {
			break
		}

		variables["cursor"] = githubv4.NewString(query.Viewer.StarredRepositories.PageInfo.EndCursor)
	}

	return allStarredRepositories, nil
}
