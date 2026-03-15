package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
)

// ContributedRepositoriesUseCase orchestrates the retrieval of contributed repositories.
type ContributedRepositoriesUseCase struct {
	commitFetcher      CommitFetcher
	issueFetcher       IssueFetcher
	pullRequestFetcher PullRequestFetcher
	repositoryFetcher  RepositoryFetcher
}

// NewContributedRepositoriesUseCase creates a new instance of ContributedRepositoriesUseCase.
func NewContributedRepositoriesUseCase(
	commitFetcher CommitFetcher,
	issueFetcher IssueFetcher,
	pullRequestFetcher PullRequestFetcher,
	repositoryFetcher RepositoryFetcher,
) *ContributedRepositoriesUseCase {
	return &ContributedRepositoriesUseCase{
		commitFetcher:      commitFetcher,
		issueFetcher:       issueFetcher,
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
	includeIssues bool,
	includePullRequests bool,
	recentActivityStatsHours int,
	adaptiveRecentActivityStats bool,
	includePrivate bool,
) ([]*domain.ContributedRepository, error) {
	// Map to track unique repositories and their latest activity time.
	// Key: "owner/repository".
	var repositoryActivityMap map[string]time.Time = make(map[string]time.Time)
	var repositoryLatestActivityMap map[string]*domain.ActivityItem = make(map[string]*domain.ActivityItem)

	// Map to track activity stats for each repository.
	var repositoryStatsMap map[string]*domain.ActivityStats = make(map[string]*domain.ActivityStats)

	// Determine initial time window for stats.
	var statsWindowHours int = recentActivityStatsHours

	if statsWindowHours <= 0 {
		statsWindowHours = 24 // Default to 24h if not specified.
	}

	// 0. Pre-fetch private repositories if needed.
	// This avoids multiple calls to FetchPrivateRepositories across commits, PRs, and issues logic.
	var privateRepos []*domain.Repository
	var fetchPrivateErr error
	privateRepos, fetchPrivateErr = useCase.repositoryFetcher.FetchPrivateRepositories(context)

	if fetchPrivateErr != nil {
		fmt.Printf("Warning: Failed to fetch private repositories: %v\n", fetchPrivateErr)
		// Continue execution, privateRepos will be nil/empty so loops will just skip.
	} else {
		// fmt.Printf("DEBUG: Pre-fetched %d private repositories\n", len(privateRepos))
	}

	// Fetch all activities first to support adaptive windows.
	// 1. Fetch Commits if requested.
	var allCommits []*domain.Commit

	if includeCommits {
		var commitError error
		allCommits, commitError = useCase.commitFetcher.FetchCommits(context, username, startTime, endTime)

		if commitError != nil {
			return nil, commitError
		}

		var commit *domain.Commit

		for _, commit = range allCommits {
			if commit.RepositoryName == "" {
				continue
			}

			var currentLatest time.Time = repositoryActivityMap[commit.RepositoryName]

			if commit.CommittedAt.After(currentLatest) {
				repositoryActivityMap[commit.RepositoryName] = commit.CommittedAt

				// Only use the first line of the commit message for the title.
				var title string = commit.Message

				if index := strings.Index(title, "\n"); index != -1 {
					title = title[:index]
				}

				repositoryLatestActivityMap[commit.RepositoryName] = &domain.ActivityItem{
					Type:      domain.ActivityTypeCommit,
					Title:     title,
					URL:       commit.HTMLURL,
					CreatedAt: commit.CommittedAt,
				}
			}
		}

		// 1.1 Fetch Private Commits if requested and username matches.
		if includePrivate && len(privateRepos) > 0 {
			var privateCommits []*domain.Commit
			var fetchPrivateCommitsErr error
			privateCommits, fetchPrivateCommitsErr = useCase.commitFetcher.FetchPrivateCommits(context, username, privateRepos, startTime, endTime)

			if fetchPrivateCommitsErr != nil {
				fmt.Printf("Failed to fetch private commits: %v\n", fetchPrivateCommitsErr)
			} else {
				// Merge private commits.
				var privateCommit *domain.Commit

				for _, privateCommit = range privateCommits {
					if privateCommit.RepositoryName == "" {
						continue
					}

					// Add to allCommits list for stats calculation later.
					allCommits = append(allCommits, privateCommit)

					var currentLatest time.Time = repositoryActivityMap[privateCommit.RepositoryName]

					if privateCommit.CommittedAt.After(currentLatest) {
						repositoryActivityMap[privateCommit.RepositoryName] = privateCommit.CommittedAt

						var title string = privateCommit.Message

						if index := strings.Index(title, "\n"); index != -1 {
							title = title[:index]
						}

						repositoryLatestActivityMap[privateCommit.RepositoryName] = &domain.ActivityItem{
							Type:      domain.ActivityTypeCommit,
							Title:     title,
							URL:       privateCommit.HTMLURL,
							CreatedAt: privateCommit.CommittedAt,
						}
					}
				}
			}
		}
	}

	// 2. Fetch Pull Requests if requested.
	var allPullRequests []*domain.PullRequest

	if includePullRequests {
		var pullRequestError error
		allPullRequests, pullRequestError = useCase.pullRequestFetcher.FetchPullRequests(context, username, startTime, endTime)

		if pullRequestError != nil {
			return nil, pullRequestError
		}

		var pullRequest *domain.PullRequest
		var currentLatest time.Time
		var repositoryName string

		for _, pullRequest = range allPullRequests {
			// Parsing repository name from URL or field.
			// The fetcher should populate RepositoryName.
			repositoryName = pullRequest.RepositoryName

			// If repository name is a URL, extract the full name.
			// E.g., https://api.github.com/repos/owner/repo.
			if strings.HasPrefix(repositoryName, "https://api.github.com/repos/") {
				repositoryName = strings.TrimPrefix(repositoryName, "https://api.github.com/repos/")
			}

			if repositoryName == "" {
				continue
			}

			currentLatest = repositoryActivityMap[repositoryName]

			if pullRequest.CreatedAt.After(currentLatest) {
				repositoryActivityMap[repositoryName] = pullRequest.CreatedAt
				repositoryLatestActivityMap[repositoryName] = &domain.ActivityItem{
					Type:      domain.ActivityTypePullRequest,
					Title:     pullRequest.Title,
					URL:       pullRequest.HTMLURL,
					CreatedAt: pullRequest.CreatedAt,
				}
			}
		}

		if includePrivate && len(privateRepos) > 0 {
			var privatePRs []*domain.PullRequest
			var fetchPRErr error
			privatePRs, fetchPRErr = useCase.pullRequestFetcher.FetchPrivatePullRequests(context, username, privateRepos, startTime, endTime)

			if fetchPRErr == nil {
				// Merge and process.
				allPullRequests = append(allPullRequests, privatePRs...)

				for _, pullRequest = range privatePRs {
					repositoryName = pullRequest.RepositoryName

					if repositoryName == "" {
						continue
					}

					currentLatest = repositoryActivityMap[repositoryName]

					if pullRequest.CreatedAt.After(currentLatest) {
						repositoryActivityMap[repositoryName] = pullRequest.CreatedAt
						repositoryLatestActivityMap[repositoryName] = &domain.ActivityItem{
							Type:      domain.ActivityTypePullRequest,
							Title:     pullRequest.Title,
							URL:       pullRequest.HTMLURL,
							CreatedAt: pullRequest.CreatedAt,
						}
					}
				}
			}
		}
	}

	// 3. Fetch Issues if requested.
	var allIssues []*domain.Issue

	if includeIssues {
		var issueError error
		allIssues, issueError = useCase.issueFetcher.FetchIssueActivities(context, username, startTime, endTime)

		if issueError != nil {
			return nil, issueError
		}

		var issue *domain.Issue
		var currentLatest time.Time
		var repositoryName string

		for _, issue = range allIssues {
			repositoryName = issue.RepositoryName

			if strings.HasPrefix(repositoryName, "https://api.github.com/repos/") {
				repositoryName = strings.TrimPrefix(repositoryName, "https://api.github.com/repos/")
			}

			if repositoryName == "" {
				continue
			}

			currentLatest = repositoryActivityMap[repositoryName]

			if issue.CreatedAt.After(currentLatest) {
				repositoryActivityMap[repositoryName] = issue.CreatedAt
				repositoryLatestActivityMap[repositoryName] = &domain.ActivityItem{
					Type:        domain.ActivityTypeIssue,
					Title:       issue.Title,
					URL:         issue.HTMLURL,
					CreatedAt:   issue.CreatedAt,
					IssueAction: issue.Action,
				}
			}
		}

		if includePrivate && len(privateRepos) > 0 {
			var privateIssues []*domain.Issue
			var fetchIssueErr error
			privateIssues, fetchIssueErr = useCase.issueFetcher.FetchPrivateIssueActivities(context, username, privateRepos, startTime, endTime)

			if fetchIssueErr == nil {
				allIssues = append(allIssues, privateIssues...)

				for _, issue = range privateIssues {
					repositoryName = issue.RepositoryName

					if repositoryName == "" {
						continue
					}

					currentLatest = repositoryActivityMap[repositoryName]

					if issue.CreatedAt.After(currentLatest) {
						repositoryActivityMap[repositoryName] = issue.CreatedAt
						repositoryLatestActivityMap[repositoryName] = &domain.ActivityItem{
							Type:        domain.ActivityTypeIssue,
							Title:       issue.Title,
							URL:         issue.HTMLURL,
							CreatedAt:   issue.CreatedAt,
							IssueAction: issue.Action,
						}
					}
				}
			}
		}
	}

	// 3. Process Stats per Repository.
	if recentActivityStatsHours > 0 || adaptiveRecentActivityStats {
		var repoName string
		for repoName = range repositoryActivityMap {

			// Function to calculate stats for a given window.
			var calculateStatsForWindow func(int) *domain.ActivityStats = func(windowHours int) *domain.ActivityStats {
				var currentCutoff time.Time = time.Now().Add(-time.Duration(windowHours) * time.Hour)
				var stats *domain.ActivityStats = &domain.ActivityStats{TimeWindow: windowHours}

				if includeCommits {
					var commit *domain.Commit

					for _, commit = range allCommits {
						if commit.RepositoryName == repoName && commit.CommittedAt.After(currentCutoff) {
							stats.CommitCount++
						}
					}
				}

				if includePullRequests {
					var pullRequest *domain.PullRequest

					for _, pullRequest = range allPullRequests {
						var prRepoName string = pullRequest.RepositoryName

						if strings.HasPrefix(prRepoName, "https://api.github.com/repos/") {
							prRepoName = strings.TrimPrefix(prRepoName, "https://api.github.com/repos/")
						}

						if prRepoName == repoName && pullRequest.CreatedAt.After(currentCutoff) {
							stats.PullRequestCount++
						}
					}
				}

				if includeIssues {
					var issue *domain.Issue

					for _, issue = range allIssues {
						var issueRepoName string = issue.RepositoryName

						if strings.HasPrefix(issueRepoName, "https://api.github.com/repos/") {
							issueRepoName = strings.TrimPrefix(issueRepoName, "https://api.github.com/repos/")
						}

						if issueRepoName == repoName && issue.CreatedAt.After(currentCutoff) {
							stats.IssueCount++
						}
					}
				}

				return stats
			}

			var stats *domain.ActivityStats

			if adaptiveRecentActivityStats {
				// Windows to check: current default -> Day -> Week -> Month -> Year.
				var windows []int = []int{24, 168, 720, 8760}

				// Check initial window first.
				stats = calculateStatsForWindow(statsWindowHours)

				if stats.CommitCount == 0 && stats.IssueCount == 0 && stats.PullRequestCount == 0 {
					// Try larger windows.
					var window int

					for _, window = range windows {
						if window > statsWindowHours {
							stats = calculateStatsForWindow(window)

							if stats.CommitCount > 0 || stats.IssueCount > 0 || stats.PullRequestCount > 0 {
								break
							}
						}
					}

					// If still no activity found in the largest window (Year), fallback to All Time.
					// Use 0 to represent All Time (Past Stats).
					if stats.CommitCount == 0 && stats.IssueCount == 0 && stats.PullRequestCount == 0 {
						stats = calculateStatsForWindow(0)
					}
				}
			} else {
				// Fixed window.
				stats = calculateStatsForWindow(statsWindowHours)
			}

			// Store stats if requested (even if empty, but logic usually implies showing stats).
			// If adaptive and still empty, we might choose to hide it or show "0 in past Year".
			// Requirement: "If no activity found even in the Year window, stats are hidden."
			// So only add if non-zero.
			if stats.CommitCount > 0 || stats.IssueCount > 0 || stats.PullRequestCount > 0 {
				repositoryStatsMap[repoName] = stats
			}
		}
	}

	// 4. Fetch details for each unique repository.
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
			fmt.Printf("Failed to fetch repository details for %s/%s: %v\n", owner, repositoryName, fetchError)
			// For robustness, skip.
			continue
		}

		if !includePrivate && repositoryDetails.Private {
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
			Repository:     repositoryDetails,
			ActiveAt:       activeAt,
			IsOwner:        isOwner,
			ActivityStats:  repositoryStatsMap[repositoryFullName],
			LatestActivity: repositoryLatestActivityMap[repositoryFullName],
		}

		contributedRepositories = append(contributedRepositories, contributedRepository)
	}

	// 4. Sort by ActiveAt descending.
	sort.Slice(contributedRepositories, func(indexA, indexB int) bool {
		return contributedRepositories[indexA].ActiveAt.After(contributedRepositories[indexB].ActiveAt)
	})

	// 5. Apply limit.
	if limitCount > 0 && len(contributedRepositories) > limitCount {
		contributedRepositories = contributedRepositories[:limitCount]
	}

	return contributedRepositories, nil
}
