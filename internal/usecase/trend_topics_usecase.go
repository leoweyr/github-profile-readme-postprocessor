package usecase

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
	"go.leoweyr.com/github-profile-postprocessor/internal/domain/enums"
)

// TrendTopicsUseCase orchestrates the calculation of trending topics.
type TrendTopicsUseCase struct {
	commitFetcher      CommitFetcher
	issueFetcher       IssueFetcher
	pullRequestFetcher PullRequestFetcher
	repositoryFetcher  RepositoryFetcher
}

// NewTrendTopicsUseCase creates a new instance of TrendTopicsUseCase.
func NewTrendTopicsUseCase(
	commitFetcher CommitFetcher,
	issueFetcher IssueFetcher,
	pullRequestFetcher PullRequestFetcher,
	repositoryFetcher RepositoryFetcher,
) *TrendTopicsUseCase {
	return &TrendTopicsUseCase{
		commitFetcher:      commitFetcher,
		issueFetcher:       issueFetcher,
		pullRequestFetcher: pullRequestFetcher,
		repositoryFetcher:  repositoryFetcher,
	}
}

// Execute calculates and returns the trending topics in markdown format.
func (useCase *TrendTopicsUseCase) Execute(
	context context.Context,
	username string,
	timeLimitHours int,
	highlightThresholdPercentage float64,
	limitCount int,
	includePrivate bool,
) (string, error) {
	if timeLimitHours <= 0 {
		timeLimitHours = 24
	}

	var startTime time.Time = time.Now().Add(-time.Duration(timeLimitHours) * time.Hour)
	var endTime time.Time = time.Now()

	// Helper to clean repo name.
	var cleanRepoName func(string) string = func(name string) string {
		if strings.HasPrefix(name, "https://api.github.com/repos/") {
			return strings.TrimPrefix(name, "https://api.github.com/repos/")
		}

		return name
	}

	// 1. Discovery Phase: Collect active repositories.
	// Fetch all active repositories to calculate global trends without count limits.
	var activeRepoNames map[string]bool = make(map[string]bool)
	var activePrivateRepos []*domain.Repository

	// 1.1 Public Activity Discovery (using Search/List APIs).
	// Fetch activities to discover active public repositories and calculate scores.
	// Fetching complete data supports both discovery and subsequent weighting.
	var allCommits []*domain.Commit
	var allPullRequests []*domain.PullRequest
	var allIssues []*domain.Issue

	var commitError error
	allCommits, commitError = useCase.commitFetcher.FetchCommits(context, username, startTime, endTime)

	if commitError != nil {
		return "", commitError
	}

	var c *domain.Commit

	for _, c = range allCommits {
		if c.RepositoryName != "" {
			activeRepoNames[c.RepositoryName] = true
		}
	}

	var prError error

	allPullRequests, prError = useCase.pullRequestFetcher.FetchPullRequests(context, username, startTime, endTime)

	if prError != nil {
		return "", prError
	}

	var p *domain.PullRequest

	for _, p = range allPullRequests {
		var repoName string = cleanRepoName(p.RepositoryName)

		if repoName != "" {
			activeRepoNames[repoName] = true
		}
	}

	var issueError error
	allIssues, issueError = useCase.issueFetcher.FetchIssueActivities(context, username, startTime, endTime)

	if issueError != nil {
		return "", issueError
	}

	var i *domain.Issue

	for _, i = range allIssues {
		var repoName string = cleanRepoName(i.RepositoryName)

		if repoName != "" {
			activeRepoNames[repoName] = true
		}
	}

	// 1.2 Private Activity Discovery.
	// Cache repositories to avoid re-fetching details later.
	var repoCache map[string]*domain.Repository = make(map[string]*domain.Repository)

	if includePrivate {
		var privateRepos []*domain.Repository
		var fetchPrivateErr error
		privateRepos, fetchPrivateErr = useCase.repositoryFetcher.FetchPrivateRepositories(context)

		if fetchPrivateErr != nil {
			fmt.Printf("Warning: Failed to fetch private repositories: %v\n", fetchPrivateErr)
		} else {
			var repo *domain.Repository

			for _, repo = range privateRepos {
				if !repo.PushedAt.IsZero() && repo.PushedAt.After(startTime) {
					activeRepoNames[repo.FullName] = true
					activePrivateRepos = append(activePrivateRepos, repo)
					repoCache[repo.FullName] = repo
				}
			}
		}
	}

	// 2. Hydration & Weighting Phase.
	// For each active repo, calculate its score.
	var repoScores map[string]float64 = make(map[string]float64)

	// Fetch private activities if needed.
	if includePrivate && len(activePrivateRepos) > 0 {
		var privateCommits []*domain.Commit
		var err error
		privateCommits, err = useCase.commitFetcher.FetchPrivateCommits(context, username, activePrivateRepos, startTime, endTime)

		if err == nil {
			allCommits = append(allCommits, privateCommits...)
		}

		var privatePRs []*domain.PullRequest
		privatePRs, err = useCase.pullRequestFetcher.FetchPrivatePullRequests(context, username, activePrivateRepos, startTime, endTime)

		if err == nil {
			allPullRequests = append(allPullRequests, privatePRs...)
		}

		var privateIssues []*domain.Issue
		privateIssues, err = useCase.issueFetcher.FetchPrivateIssueActivities(context, username, activePrivateRepos, startTime, endTime)

		if err == nil {
			allIssues = append(allIssues, privateIssues...)
		}
	}

	// Calculate weights.
	// Weighting Rule: Day Index * Type Weight.
	// Day Index: 1 (Oldest) to TotalDays (Newest).
	var totalDays int = int(math.Ceil(float64(timeLimitHours) / 24.0))

	var calculateDayIndex func(t time.Time) int = func(t time.Time) int {
		var durationSinceStart time.Duration = t.Sub(startTime)
		var hoursSinceStart float64 = durationSinceStart.Hours()

		if hoursSinceStart < 0 {
			return 0
		}

		var dayIndex int = int(math.Floor(hoursSinceStart/24.0)) + 1

		if dayIndex > totalDays {
			dayIndex = totalDays
		}

		return dayIndex
	}

	// Process Commits (Weight: 1).
	for _, c = range allCommits {
		if c.RepositoryName == "" {
			continue
		}

		var dayIndex int = calculateDayIndex(c.CommittedAt)
		repoScores[c.RepositoryName] += float64(dayIndex) * 1.0
	}

	// Process PRs (Weight: 1).
	for _, p = range allPullRequests {
		var repoName string = cleanRepoName(p.RepositoryName)

		if repoName == "" {
			continue
		}

		var dayIndex int = calculateDayIndex(p.CreatedAt)
		repoScores[repoName] += float64(dayIndex) * 1.0
	}

	// Process Issues (Weight: 1).
	for _, i = range allIssues {
		var repoName string = cleanRepoName(i.RepositoryName)

		if repoName == "" {
			continue
		}

		// Only count 'created' or 'commented' actions (already filtered by fetcher logic essentially).
		// IssueFetcher returns both created and commented issues.
		if i.Action == enums.IssueActionCreated || i.Action == enums.IssueActionCommented {
			var dayIndex int = calculateDayIndex(i.CreatedAt)
			repoScores[repoName] += float64(dayIndex) * 1.0
		}
	}

	// 3. Topic Aggregation.
	var topicScores map[string]float64 = make(map[string]float64)
	var totalTopicScore float64 = 0

	var repoName string
	var score float64

	for repoName, score = range repoScores {
		if score == 0 {
			continue
		}

		// Fetch repo details to get topics.
		// Check cache first (populated by private repo discovery).
		var repoDetails *domain.Repository = repoCache[repoName]

		if repoDetails == nil {
			var parts []string = strings.Split(repoName, "/")

			if len(parts) != 2 {
				continue
			}

			var err error
			repoDetails, err = useCase.repositoryFetcher.FetchRepository(context, parts[0], parts[1])

			if err != nil {
				fmt.Printf("Warning: Failed to fetch details for %s: %v\n", repoName, err)

				continue
			}

			// Cache for future reference (though this is end of flow).
			repoCache[repoName] = repoDetails
		}

		if !includePrivate && repoDetails.Private {
			continue
		}

		var topic string

		for _, topic = range repoDetails.Topics {
			var normalizedTopic string = strings.ToLower(strings.TrimSpace(topic))

			if normalizedTopic == "" {
				continue
			}

			topicScores[normalizedTopic] += score
			totalTopicScore += score
		}
	}

	// 4. Sorting & Formatting.
	var trendTopics []*domain.TrendTopic
	var name string

	for name, score = range topicScores {
		var percentage float64 = 0

		if totalTopicScore > 0 {
			percentage = score / totalTopicScore
		}

		trendTopics = append(trendTopics, &domain.TrendTopic{
			Name:       name,
			Score:      score,
			Percentage: percentage,
		})
	}

	sort.Slice(trendTopics, func(i, j int) bool {
		return trendTopics[i].Score > trendTopics[j].Score
	})

	if limitCount > 0 && len(trendTopics) > limitCount {
		trendTopics = trendTopics[:limitCount]
	}

	// Format Output.
	var resultParts []string
	var topic *domain.TrendTopic

	for _, topic = range trendTopics {
		if topic.Percentage >= highlightThresholdPercentage {
			// Highlight: code(percentage).
			resultParts = append(resultParts, fmt.Sprintf("<code>%s(%.0f%%)</code>", topic.Name, topic.Percentage*100))
		} else {
			// Normal: code.
			resultParts = append(resultParts, fmt.Sprintf("<code>%s</code>", topic.Name))
		}
	}

	return strings.Join(resultParts, " "), nil
}

func cleanRepoName(name string) string {
	if strings.HasPrefix(name, "https://api.github.com/repos/") {
		return strings.TrimPrefix(name, "https://api.github.com/repos/")
	}

	return name
}
