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
	// Optimization Strategy (Ladder Algorithm):
	// 1. Discovery Phase: Collect all active repository candidates.
	//    - Public: Fetch active repos via Search API (efficient).
	//    - Private: Fetch all private repos and use PushedAt as initial sorting key (avoids deep fetching).
	// 2. Ranking Phase: Sort candidates by activity time and slice the Top N (limitCount).
	// 3. Hydration Phase: For only the Top N repositories:
	//    - Fetch precise "Latest Activity" details (Commit/PR/Issue) if missing (Private).
	//    - Calculate activity stats using adaptive time windows (Ladder) to minimize API calls.

	// --- Phase 1: Candidate Discovery ---

	// Map to track unique repositories and their latest activity time.
	// Key: "owner/repository".
	var repositoryActivityMap map[string]time.Time = make(map[string]time.Time)
	var repositoryLatestActivityMap map[string]*domain.ActivityItem = make(map[string]*domain.ActivityItem)
	// Track strict ownership for later stats hydration.
	var repositoryIsPrivateMap map[string]*domain.Repository = make(map[string]*domain.Repository)

	// 1.1 Public Activity Discovery (using Search API).
	// Retain existing logic for public data due to Search efficiency.
	var allCommits []*domain.Commit
	var allPullRequests []*domain.PullRequest
	var allIssues []*domain.Issue

	if includeCommits {
		var commitError error
		allCommits, commitError = useCase.commitFetcher.FetchCommits(context, username, startTime, endTime)

		if commitError != nil {
			return nil, commitError
		}

		var commit *domain.Commit

		for _, commit = range allCommits {
			if commit == nil {
				continue
			}

			if commit.RepositoryName == "" {
				continue
			}

			var currentLatest time.Time = repositoryActivityMap[commit.RepositoryName]

			if commit.CommittedAt.After(currentLatest) {
				repositoryActivityMap[commit.RepositoryName] = commit.CommittedAt
				var title string = commit.Message

				var index int = strings.Index(title, "\n")

				if index != -1 {
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
	}

	if includePullRequests {
		var pullRequestError error
		allPullRequests, pullRequestError = useCase.pullRequestFetcher.FetchPullRequests(context, username, startTime, endTime)

		if pullRequestError != nil {
			return nil, pullRequestError
		}
		var pullRequest *domain.PullRequest

		for _, pullRequest = range allPullRequests {
			if pullRequest == nil {
				continue
			}

			var repoName string = pullRequest.RepositoryName

			if strings.HasPrefix(repoName, "https://api.github.com/repos/") {
				repoName = strings.TrimPrefix(repoName, "https://api.github.com/repos/")
			}

			if repoName == "" {
				continue
			}

			var currentLatest time.Time = repositoryActivityMap[repoName]

			if pullRequest.CreatedAt.After(currentLatest) {
				repositoryActivityMap[repoName] = pullRequest.CreatedAt
				repositoryLatestActivityMap[repoName] = &domain.ActivityItem{
					Type:      domain.ActivityTypePullRequest,
					Title:     pullRequest.Title,
					URL:       pullRequest.HTMLURL,
					CreatedAt: pullRequest.CreatedAt,
				}
			}
		}
	}

	if includeIssues {
		var issueError error
		allIssues, issueError = useCase.issueFetcher.FetchIssueActivities(context, username, startTime, endTime)

		if issueError != nil {
			return nil, issueError
		}
		var issue *domain.Issue

		for _, issue = range allIssues {
			if issue == nil {
				continue
			}

			var repoName string = issue.RepositoryName

			if strings.HasPrefix(repoName, "https://api.github.com/repos/") {
				repoName = strings.TrimPrefix(repoName, "https://api.github.com/repos/")
			}

			if repoName == "" {
				continue
			}

			var currentLatest time.Time = repositoryActivityMap[repoName]

			if issue.CreatedAt.After(currentLatest) {
				repositoryActivityMap[repoName] = issue.CreatedAt
				repositoryLatestActivityMap[repoName] = &domain.ActivityItem{
					Type:        domain.ActivityTypeIssue,
					Title:       issue.Title,
					URL:         issue.HTMLURL,
					CreatedAt:   issue.CreatedAt,
					IssueAction: issue.Action,
				}
			}
		}
	}

	// 1.2 Private Activity Discovery (Optimized).
	if includePrivate {
		var privateRepos []*domain.Repository
		var fetchPrivateErr error
		privateRepos, fetchPrivateErr = useCase.repositoryFetcher.FetchPrivateRepositories(context)

		if fetchPrivateErr != nil {
			fmt.Printf("Warning: Failed to fetch private repositories: %v\n", fetchPrivateErr)
		} else {
			var repo *domain.Repository

			for _, repo = range privateRepos {
				if repo == nil {
					continue
				}

				// Use PushedAt as a lightweight proxy for "Latest Activity".
				// This avoids fetching commits/issues for every private repo.
				if !repo.PushedAt.IsZero() && !repo.PushedAt.Before(startTime) {
					var currentLatest time.Time = repositoryActivityMap[repo.FullName]

					if repo.PushedAt.After(currentLatest) {
						repositoryActivityMap[repo.FullName] = repo.PushedAt
						// Mark as private to trigger detailed fetching if selected.
						repositoryIsPrivateMap[repo.FullName] = repo
					}
				}
			}
		}
	}

	// --- Phase 2: Ranking & Limiting ---
	type Candidate struct {
		RepoName string
		ActiveAt time.Time
	}

	var candidates []Candidate

	var name string
	var activeAt time.Time

	for name, activeAt = range repositoryActivityMap {
		candidates = append(candidates, Candidate{RepoName: name, ActiveAt: activeAt})
	}

	// Sort by ActiveAt descending.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ActiveAt.After(candidates[j].ActiveAt)
	})

	// Slice Top N logic resides inside the loop to support filtering.
	// Iterate through all candidates until finding enough matching repositories.
	var contributedRepositories []*domain.ContributedRepository
	var repositoryStatsMap map[string]*domain.ActivityStats = make(map[string]*domain.ActivityStats)

	var candidate Candidate

	for _, candidate = range candidates {
		// Stop if we have reached the requested limit.
		if limitCount > 0 && len(contributedRepositories) >= limitCount {
			break
		}

		var repoName string = candidate.RepoName
		var repoDetails *domain.Repository
		var fetchError error

		// Fetch basic details (Public or Private).
		var parts []string = strings.Split(repoName, "/")

		if len(parts) != 2 {
			continue
		}

		repoDetails, fetchError = useCase.repositoryFetcher.FetchRepository(context, parts[0], parts[1])

		if fetchError != nil {
			fmt.Printf("Warning: Failed to fetch details for %s: %v\n", repoName, fetchError)
			continue
		}

		// Apply Filters (Name/Topic).
		var hasRepositoryNameFilters bool = len(repositoryNameFilters) > 0
		var hasRepositoryTopicFilters bool = len(repositoryTopicFilters) > 0
		var repositoryNameMatch bool = false

		if hasRepositoryNameFilters {
			var filter string

			for _, filter = range repositoryNameFilters {
				if filter != "" && strings.Contains(strings.ToLower(repoDetails.FullName), strings.ToLower(filter)) {
					repositoryNameMatch = true

					break
				}
			}
		}

		if !includePrivate && repoDetails.Private {
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

				for _, topic = range repoDetails.Topics {
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

		// Check if it's a private repo that needs hydration.
		var privateRepo *domain.Repository = repositoryIsPrivateMap[repoName]
		var isPrivateCandidate bool = privateRepo != nil

		// 3.1 Hydrate Latest Activity (for Private only).
		// Public repos already have `repositoryLatestActivityMap` populated from Phase 1.
		if isPrivateCandidate {
			// Determine the actual latest event to display the correct Title, URL, and Type.
			// PushedAt was just a proxy.
			var latestCommit *domain.Commit
			var latestPR *domain.PullRequest
			var latestIssue *domain.Issue
			var err error

			// Fetch latest of each type (limit 1).
			if includeCommits {
				latestCommit, err = useCase.commitFetcher.GetLatestPrivateCommit(context, username, privateRepo)

				if err != nil {
					fmt.Printf("Debug: Failed to get latest commit for %s: %v\n", repoName, err)
				}
			}

			if includePullRequests {
				latestPR, err = useCase.pullRequestFetcher.GetLatestPrivatePullRequest(context, username, privateRepo)

				if err != nil {
					fmt.Printf("Debug: Failed to get latest PR for %s: %v\n", repoName, err)
				}
			}

			if includeIssues {
				latestIssue, err = useCase.issueFetcher.GetLatestPrivateIssue(context, username, privateRepo)

				if err != nil {
					fmt.Printf("Debug: Failed to get latest Issue for %s: %v\n", repoName, err)
				}
			}

			// Compare to find the winner.
			var bestTime time.Time
			var bestItem *domain.ActivityItem

			if latestCommit != nil && latestCommit.CommittedAt.After(bestTime) {
				bestTime = latestCommit.CommittedAt
				var title string = latestCommit.Message
				var idx int = strings.Index(title, "\n")

				if idx != -1 {
					title = title[:idx]
				}
				bestItem = &domain.ActivityItem{Type: domain.ActivityTypeCommit, Title: title, URL: latestCommit.HTMLURL, CreatedAt: bestTime}
			}

			if latestPR != nil && latestPR.CreatedAt.After(bestTime) {
				bestTime = latestPR.CreatedAt
				bestItem = &domain.ActivityItem{Type: domain.ActivityTypePullRequest, Title: latestPR.Title, URL: latestPR.HTMLURL, CreatedAt: bestTime}
			}

			if latestIssue != nil && latestIssue.CreatedAt.After(bestTime) {
				bestTime = latestIssue.CreatedAt
				bestItem = &domain.ActivityItem{Type: domain.ActivityTypeIssue, Title: latestIssue.Title, URL: latestIssue.HTMLURL, CreatedAt: bestTime, IssueAction: latestIssue.Action}
			}

			if bestItem != nil {
				repositoryLatestActivityMap[repoName] = bestItem
				// Update ActiveAt to precise time.
				candidate.ActiveAt = bestTime
			}

			// Critical: If no latest activity exists for the user after checking all types,
			// it means the repository was picked up by PushedAt (Phase 1.2) but the user hasn't actually contributed,
			// or the user contributed outside the scope/time.
			// Filter out these false positives.
			if repositoryLatestActivityMap[repoName] == nil {
				// No activity found for this user in this private repo. Skip it.
				continue
			}
		}

		// 3.2 Hydrate Stats (Ladder Strategy).
		var windows []int = []int{recentActivityStatsHours}

		if adaptiveRecentActivityStats {
			windows = []int{24, 168, 720, 8760} // Day, Week, Month, Year.
		}

		if recentActivityStatsHours <= 0 && !adaptiveRecentActivityStats {
			windows = []int{24} // Default.
		}

		var stats *domain.ActivityStats
		var window int

		for _, window = range windows {
			var currentCutoff time.Time = time.Now().Add(-time.Duration(window) * time.Hour)
			var currentStats *domain.ActivityStats = &domain.ActivityStats{TimeWindow: window}

			// Public Stats: Count in-memory from Phase 1 data.
			if !isPrivateCandidate {
				if includeCommits {
					var c *domain.Commit

					for _, c = range allCommits {
						if c == nil {
							continue
						}

						if c.RepositoryName == repoName && c.CommittedAt.After(currentCutoff) {
							currentStats.CommitCount++
						}
					}
				}

				if includePullRequests {
					var p *domain.PullRequest

					for _, p = range allPullRequests {
						if p == nil {
							continue
						}

						var pRepo string = p.RepositoryName

						if strings.HasPrefix(pRepo, "https://api.github.com/repos/") {
							pRepo = strings.TrimPrefix(pRepo, "https://api.github.com/repos/")
						}

						if pRepo == repoName && p.CreatedAt.After(currentCutoff) {
							currentStats.PullRequestCount++
						}
					}
				}

				if includeIssues {
					var i *domain.Issue

					for _, i = range allIssues {
						if i == nil {
							continue
						}

						var iRepo string = i.RepositoryName

						if strings.HasPrefix(iRepo, "https://api.github.com/repos/") {
							iRepo = strings.TrimPrefix(iRepo, "https://api.github.com/repos/")
						}

						if iRepo == repoName && i.CreatedAt.After(currentCutoff) {
							currentStats.IssueCount++
						}
					}
				}
			} else {
				// Private Stats: Fetch Count via API (Ladder).
				if includeCommits {
					var count int
					var err error

					// Note: privateRepo is guaranteed non-nil here because isPrivateCandidate is checked above
					if privateRepo != nil {
						count, err = useCase.commitFetcher.CountPrivateCommits(context, username, privateRepo, currentCutoff, time.Now())
					}

					if err == nil {
						currentStats.CommitCount = count
					}
				}

				if includePullRequests {
					var count int
					var err error

					if privateRepo != nil {
						count, err = useCase.pullRequestFetcher.CountPrivatePullRequests(context, username, privateRepo, currentCutoff, time.Now())
					}

					if err == nil {
						currentStats.PullRequestCount = count
					}
				}

				if includeIssues {
					var count int
					var err error

					if privateRepo != nil {
						count, err = useCase.issueFetcher.CountPrivateIssues(context, username, privateRepo, currentCutoff, time.Now())
					}

					if err == nil {
						currentStats.IssueCount = count
					}
				}
			}

			// If activity found, accept this window and stop laddering.
			if currentStats.CommitCount > 0 || currentStats.PullRequestCount > 0 || currentStats.IssueCount > 0 {
				stats = currentStats

				break
			}

			// Set empty stats if count remains zero in the final window.
		}

		if stats != nil {
			repositoryStatsMap[repoName] = stats
		}

		// Build Result.
		var isOwner bool = (strings.EqualFold(repoDetails.Owner, username))
		var contributedRepository *domain.ContributedRepository = &domain.ContributedRepository{
			Repository:     repoDetails,
			ActiveAt:       candidate.ActiveAt,
			IsOwner:        isOwner,
			ActivityStats:  repositoryStatsMap[repoName],
			LatestActivity: repositoryLatestActivityMap[repoName],
		}

		contributedRepositories = append(contributedRepositories, contributedRepository)
	}

	return contributedRepositories, nil
}
