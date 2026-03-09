package filter

import (
	"strings"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// RepositoryFilter provides methods to filter a list of repositories based on specific criteria.
type RepositoryFilter struct {
	repositories []*domain.Repository
}

// NewRepositoryFilter creates a new instance of RepositoryFilter with the initial list of repositories.
func NewRepositoryFilter(repositories []*domain.Repository) *RepositoryFilter {
	return &RepositoryFilter{
		repositories: repositories,
	}
}

// FilterByName retains only the repositories whose names contain the specified substring, ignoring case.
func (filter *RepositoryFilter) FilterByName(nameSubstring string) *RepositoryFilter {
	var filteredRepositories []*domain.Repository
	var lowerCaseSubstring string = strings.ToLower(nameSubstring)

	for _, repository := range filter.repositories {
		var lowerCaseName string = strings.ToLower(repository.Name)

		if strings.Contains(lowerCaseName, lowerCaseSubstring) {
			filteredRepositories = append(filteredRepositories, repository)
		}
	}

	filter.repositories = filteredRepositories

	return filter
}

// FilterByTopic retains only the repositories that have at least one topic containing the specified substring, ignoring case.
func (filter *RepositoryFilter) FilterByTopic(topicSubstring string) *RepositoryFilter {
	var filteredRepositories []*domain.Repository
	var lowerCaseSubstring string = strings.ToLower(topicSubstring)

	for _, repository := range filter.repositories {
		var hasMatchingTopic bool = false

		for _, topic := range repository.Topics {
			var lowerCaseTopic string = strings.ToLower(topic)

			if strings.Contains(lowerCaseTopic, lowerCaseSubstring) {
				hasMatchingTopic = true
				break
			}
		}

		if hasMatchingTopic {
			filteredRepositories = append(filteredRepositories, repository)
		}
	}

	filter.repositories = filteredRepositories

	return filter
}

// GetResult returns the current list of repositories held by the filter.
func (filter *RepositoryFilter) GetResult() []*domain.Repository {
	return filter.repositories
}
