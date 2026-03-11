package fetcher

import (
	"context"
	"fmt"
	"net/http"
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
	var httpClient *http.Client = oauth2.NewClient(context.Background(), tokenSource)
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
						Name             string
						NameWithOwner    string
						Url              string
						RepositoryTopics struct {
							Nodes []struct {
								Topic struct {
									Name string
								}
							}
						} `graphql:"repositoryTopics(first: 10)"`
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
		var queryError error = fetcher.client.Query(context, &query, variables)

		if queryError != nil {
			return nil, fmt.Errorf("failed to query starred repositories: %w", queryError)
		}

		var edge struct {
			StarredAt time.Time
			Node      struct {
				Name             string
				NameWithOwner    string
				Url              string
				RepositoryTopics struct {
					Nodes []struct {
						Topic struct {
							Name string
						}
					}
				} `graphql:"repositoryTopics(first: 10)"`
			}
		}

		for _, edge = range query.Viewer.StarredRepositories.Edges {
			var starredAt time.Time = edge.StarredAt

			// Skip the star if it is newer than the end time.
			if starredAt.After(endTime) {
				continue
			}

			// Stop fetching if the star is older than the start time.
			if starredAt.Before(startTime) {
				return allStarredRepositories, nil
			}

			var topics []string

			var topicNode struct {
				Topic struct {
					Name string
				}
			}

			for _, topicNode = range edge.Node.RepositoryTopics.Nodes {
				topics = append(topics, topicNode.Topic.Name)
			}

			var starredRepository *domain.Repository = &domain.Repository{
				Name:      edge.Node.Name,
				FullName:  edge.Node.NameWithOwner,
				HTMLURL:   edge.Node.Url,
				StarredAt: starredAt,
				Topics:    topics,
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
