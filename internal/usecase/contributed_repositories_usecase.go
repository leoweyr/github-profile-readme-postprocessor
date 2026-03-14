package usecase

import (
	"context"
	"sort"
	"strings"
	"time"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// ContributedRepositoriesUseCase orchestrates the retrieval of contributed repositories.
type ContributedRepositoriesUseCase struct {
	commitFetcher      CommitFetcher
	pullRequestFetcher PullRequestFetcher
	repositoryFetcher  RepositoryFetcher
}

// NewContributedRepositoriesUseCase creates a new instance of ContributedRepositoriesUseCase.
func NewContributedRepositoriesUseCase(
	commitFetcher CommitFetcher,
	pullRequestFetcher PullRequestFetcher,
	repositoryFetcher RepositoryFetcher,
) *ContributedRepositoriesUseCase {
	return &ContributedRepositoriesUseCase{
		commitFetcher:      commitFetcher,
		pullRequestFetcher: pullRequestFetcher,
		repositoryFetcher:  repositoryFetcher,
	}
}

// Execute retrieves and filters contributed repositories based on the provided criteria.
func (useCase *ContributedRepositoriesUseCase) Execute(
	context context.Context,
	username string,
	startTime time.Time,
	endTime time.Time,
	limitCount int,
	repositoryNameFilters []string,
	repositoryTopicFilters []string,
	includeCommits bool,
	includePullRequests bool,
	recentActivityStatsHours int,
) ([]*domain.ContributedRepository, error) {
	// Map to track unique repositories and their latest activity time.
	// Key: "owner/repository".
	var repositoryActivityMap map[string]time.Time = make(map[string]time.Time)
	var repositoryStatsMap map[string]*domain.ActivityStats = make(map[string]*domain.ActivityStats)

	// Define stats cutoff time.
	var statsCutoffTime time.Time
	if recentActivityStatsHours > 0 {
		statsCutoffTime = time.Now().Add(-time.Duration(recentActivityStatsHours) * time.Hour)
	}

	// 1. Fetch Commits if requested.
	if includeCommits {
		var commits []*domain.Commit
		var commitError error
		commits, commitError = useCase.commitFetcher.FetchCommits(context, username, startTime, endTime)

		if commitError != nil {
			return nil, commitError
		}

		var commit *domain.Commit

		for _, commit = range commits {
			if commit.RepositoryName == "" {
				continue
			}

			var currentLatest time.Time = repositoryActivityMap[commit.RepositoryName]

			if commit.CommittedAt.After(currentLatest) {
				repositoryActivityMap[commit.RepositoryName] = commit.CommittedAt
			}

			// Calculate stats.
			if recentActivityStatsHours > 0 && commit.CommittedAt.After(statsCutoffTime) {
				if repositoryStatsMap[commit.RepositoryName] == nil {
					repositoryStatsMap[commit.RepositoryName] = &domain.ActivityStats{}
				}

				repositoryStatsMap[commit.RepositoryName].CommitCount++
			}
		}
	}

	// 2. Fetch Pull Requests if requested.
	if includePullRequests {
		var pullRequests []*domain.PullRequest
		var pullRequestError error
		pullRequests, pullRequestError = useCase.pullRequestFetcher.FetchPullRequests(context, username, startTime, endTime)

		if pullRequestError != nil {
			return nil, pullRequestError
		}

		var pullRequest *domain.PullRequest

		for _, pullRequest = range pullRequests {
			// Parsing repository name from URL or field.
			// The fetcher should populate RepositoryName.
			var repositoryName string = pullRequest.RepositoryName

			// If repository name is a URL, extract the full name.
			// E.g., https://api.github.com/repos/owner/repo.
			if strings.HasPrefix(repositoryName, "https://api.github.com/repos/") {
				repositoryName = strings.TrimPrefix(repositoryName, "https://api.github.com/repos/")
			}

			if repositoryName == "" {
				continue
			}

			var currentLatest time.Time = repositoryActivityMap[repositoryName]

			if pullRequest.CreatedAt.After(currentLatest) {
				repositoryActivityMap[repositoryName] = pullRequest.CreatedAt
			}

			// Calculate stats.
			if recentActivityStatsHours > 0 && pullRequest.CreatedAt.After(statsCutoffTime) {
				if repositoryStatsMap[repositoryName] == nil {
					repositoryStatsMap[repositoryName] = &domain.ActivityStats{}
				}

				repositoryStatsMap[repositoryName].PullRequestCount++
			}
		}
	}

	// 3. Fetch details for each unique repository.
	var contributedRepositories []*domain.ContributedRepository

	var repositoryFullName string
	var activeAt time.Time

	for repositoryFullName, activeAt = range repositoryActivityMap {
		var parts []string = strings.Split(repositoryFullName, "/")

		if len(parts) != 2 {
			continue
		}

		var owner string = parts[0]

		var repositoryName string = parts[1]

		var hasRepositoryNameFilters bool = len(repositoryNameFilters) > 0
		var hasRepositoryTopicFilters bool = len(repositoryTopicFilters) > 0
		var repositoryNameMatch bool = false

		if hasRepositoryNameFilters {
			var filter string

			for _, filter = range repositoryNameFilters {
				if filter != "" && strings.Contains(strings.ToLower(repositoryFullName), strings.ToLower(filter)) {
					repositoryNameMatch = true
					break
				}
			}
		}

		var repositoryDetails *domain.Repository
		var fetchError error
		repositoryDetails, fetchError = useCase.repositoryFetcher.FetchRepository(context, owner, repositoryName)

		if fetchError != nil {
			// If fetching details fails, we might skip this repo or log error.
			// For robustness, skip.
			continue
		}

		var repositoryTopicMatch bool = false

		if hasRepositoryTopicFilters {
			var filter string

			for _, filter = range repositoryTopicFilters {
				if filter == "" {
					continue
				}

				var topic string

				for _, topic = range repositoryDetails.Topics {
					if strings.Contains(strings.ToLower(topic), strings.ToLower(filter)) {
						repositoryTopicMatch = true
						break
					}
				}

				if repositoryTopicMatch {
					break
				}
			}
		}

		if hasRepositoryNameFilters && hasRepositoryTopicFilters {
			if !repositoryNameMatch && !repositoryTopicMatch {
				continue
			}
		} else if hasRepositoryNameFilters {
			if !repositoryNameMatch {
				continue
			}
		} else if hasRepositoryTopicFilters {
			if !repositoryTopicMatch {
				continue
			}
		}

		var isOwner bool = (strings.EqualFold(repositoryDetails.Owner, username))

		var contributedRepository *domain.ContributedRepository = &domain.ContributedRepository{
			Repository:    repositoryDetails,
			ActiveAt:      activeAt,
			IsOwner:       isOwner,
			ActivityStats: repositoryStatsMap[repositoryFullName],
		}

		contributedRepositories = append(contributedRepositories, contributedRepository)
	}

	// 4. Sort by ActiveAt descending.
	sort.Slice(contributedRepositories, func(i, j int) bool {
		return contributedRepositories[i].ActiveAt.After(contributedRepositories[j].ActiveAt)
	})

	// 5. Apply limit.
	if limitCount > 0 && len(contributedRepositories) > limitCount {
		contributedRepositories = contributedRepositories[:limitCount]
	}

	return contributedRepositories, nil
}
