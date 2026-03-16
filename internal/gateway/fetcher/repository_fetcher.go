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

// RepositoryFetcher handles the retrieval of repository details.
type RepositoryFetcher struct {
	client *github.Client
}

// NewRepositoryFetcher creates a new instance of RepositoryFetcher with the provided token.
func NewRepositoryFetcher(token string) *RepositoryFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	var httpClient *http.Client = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &RepositoryFetcher{
		client: client,
	}
}

// FetchRepository retrieves details for a specific repository.
func (fetcher *RepositoryFetcher) FetchRepository(context context.Context, owner, repositoryName string) (*domain.Repository, error) {
	var repository *github.Repository
	var response *github.Response
	var fetchError error

	repository, response, fetchError = fetcher.client.Repositories.Get(context, owner, repositoryName)

	if fetchError != nil {
		return nil, fmt.Errorf("failed to fetch repository %s/%s: %w", owner, repositoryName, fetchError)
	}

	if repository == nil {
		return nil, fmt.Errorf("repository %s/%s not found (nil response)", owner, repositoryName)
	}

	var description string

	if repository.Description != nil {
		description = *repository.Description
	}

	var topics []string = repository.Topics

	var ownerName string

	if repository.Owner != nil && repository.Owner.Login != nil {
		ownerName = repository.GetOwner().GetLogin()
	}

	var domainRepository *domain.Repository = &domain.Repository{
		Name:        repository.GetName(),
		FullName:    repository.GetFullName(),
		Description: description,
		HTMLURL:     repository.GetHTMLURL(),
		Owner:       ownerName,
		Topics:      topics,
		Private:     repository.GetPrivate(),
	}

	// Rate limit check could be added here if needed using response.
	_ = response

	return domainRepository, nil
}

// FetchPrivateRepositories retrieves a list of private repositories for the authenticated user.
func (fetcher *RepositoryFetcher) FetchPrivateRepositories(context context.Context) ([]*domain.Repository, error) {
	var allPrivateRepos []*domain.Repository
	var listOptions *github.RepositoryListOptions = &github.RepositoryListOptions{
		Visibility:  "private",
		Affiliation: "owner,collaborator,organization_member", // Include repos where user is owner, collaborator, or org member.
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		var repositories []*github.Repository
		var response *github.Response
		var listError error
		repositories, response, listError = fetcher.client.Repositories.List(context, "", listOptions)

		if listError != nil {
			return nil, fmt.Errorf("failed to list private repositories: %w", listError)
		}

		var repository *github.Repository

		for _, repository = range repositories {
			if repository == nil {
				continue
			}

			var pushedAt time.Time

			if repository.PushedAt != nil {
				pushedAt = repository.PushedAt.Time
			}

			var ownerName string

			if repository.Owner != nil && repository.Owner.Login != nil {
				ownerName = repository.GetOwner().GetLogin()
			}

			var domainRepository *domain.Repository = &domain.Repository{
				Name:        repository.GetName(),
				FullName:    repository.GetFullName(),
				Description: repository.GetDescription(),
				HTMLURL:     repository.GetHTMLURL(),
				Owner:       ownerName,
				Topics:      repository.Topics,
				Private:     true,
				PushedAt:    pushedAt,
			}

			allPrivateRepos = append(allPrivateRepos, domainRepository)
		}

		if response.NextPage == 0 {
			break
		}

		listOptions.Page = response.NextPage
	}

	return allPrivateRepos, nil
}
