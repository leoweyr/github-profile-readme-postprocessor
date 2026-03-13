package fetcher

import (
	"context"
	"fmt"
	"net/http"

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

	var description string

	if repository.Description != nil {
		description = *repository.Description
	}

	var topics []string = repository.Topics

	var domainRepository *domain.Repository = &domain.Repository{
		Name:        repository.GetName(),
		FullName:    repository.GetFullName(),
		Description: description,
		HTMLURL:     repository.GetHTMLURL(),
		Owner:       repository.GetOwner().GetLogin(),
		Topics:      topics,
	}

	// Rate limit check could be added here if needed using response.
	_ = response

	return domainRepository, nil
}
