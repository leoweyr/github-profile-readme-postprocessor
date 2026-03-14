package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.leoweyr.com/github-profile-postprocessor/internal/delivery/model"
	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
	"go.leoweyr.com/github-profile-postprocessor/internal/usecase"
)

// ContributedRepositoriesController handles HTTP requests related to GitHub repositories.
type ContributedRepositoriesController struct {
	useCase *usecase.ContributedRepositoriesUseCase
}

// NewContributedRepositoriesController creates a new instance of ContributedRepositoriesController.
func NewContributedRepositoriesController(useCase *usecase.ContributedRepositoriesUseCase) *ContributedRepositoriesController {
	return &ContributedRepositoriesController{
		useCase: useCase,
	}
}

// parseQueryParameters extracts and validates common query parameters.
func (controller *ContributedRepositoriesController) parseQueryParameters(request *http.Request) (
	username string,
	limitCount int,
	startTime time.Time,
	endTime time.Time,
	repositoryNameFilters []string,
	repositoryTopicFilters []string,
	includeCommits bool,
	includePullRequests bool,
	includeIssues bool,
	showRecentActivityStats int,
	adaptiveRecentActivityStats bool,
	showLatestActivity bool,
	parseError error,
) {
	var queryValues url.Values = request.URL.Query()

	username = queryValues.Get("username")

	if username == "" {
		return "", 0, time.Time{}, time.Time{}, nil, nil, false, false, false, 0, false, false, fmt.Errorf("missing required parameter: username")
	}

	limitCount = 3
	var value string = queryValues.Get("limit_count")

	if value != "" {
		var parsed int
		var conversionError error
		parsed, conversionError = strconv.Atoi(value)

		if conversionError == nil && parsed > 0 {
			limitCount = parsed
		}
	}

	// Default time range: last 30 days if not specified.
	endTime = time.Now()
	value = queryValues.Get("until")

	if value != "" {
		var parsed time.Time
		var parseError error
		parsed, parseError = time.Parse(time.RFC3339, value)

		if parseError == nil {
			endTime = parsed
		}
	}

	startTime = endTime.AddDate(0, 0, -30)
	value = queryValues.Get("since")

	if value != "" {
		var parsed time.Time
		var parseError error
		parsed, parseError = time.Parse(time.RFC3339, value)

		if parseError == nil {
			startTime = parsed
		}
	}

	var rawNameFilters []string = queryValues["repository_name_contains_any"]
	repositoryNameFilters = make([]string, 0)

	var rawFilter string

	for _, rawFilter = range rawNameFilters {
		var cleanRaw string = strings.Trim(rawFilter, "\"")
		var splits []string = strings.Split(cleanRaw, ",")

		var splitPart string

		for _, splitPart = range splits {
			var trimmed string = strings.Trim(splitPart, "\"")
			trimmed = strings.TrimSpace(trimmed)

			if trimmed != "" {
				repositoryNameFilters = append(repositoryNameFilters, trimmed)
			}
		}
	}

	var rawTopicFilters []string = queryValues["repository_topic_contains_any"]
	repositoryTopicFilters = make([]string, 0)

	for _, rawFilter = range rawTopicFilters {
		var cleanRaw string = strings.Trim(rawFilter, "\"")
		var splits []string = strings.Split(cleanRaw, ",")

		var splitPart string

		for _, splitPart = range splits {
			var trimmed string = strings.Trim(splitPart, "\"")
			trimmed = strings.TrimSpace(trimmed)
			if trimmed != "" {
				repositoryTopicFilters = append(repositoryTopicFilters, trimmed)
			}
		}
	}

	includeCommits = true
	value = queryValues.Get("include_commits")

	if value == "false" {
		includeCommits = false
	}

	includePullRequests = true
	value = queryValues.Get("include_pull_requests")

	if value == "false" {
		includePullRequests = false
	}

	includeIssues = true
	value = queryValues.Get("include_issues")

	if value == "false" {
		includeIssues = false
	}

	showRecentActivityStats = 0
	value = queryValues.Get("show_recent_activity_stats")

	if value != "" {
		var parsed int
		var conversionError error
		parsed, conversionError = strconv.Atoi(value)
		if conversionError == nil && parsed >= 0 {
			showRecentActivityStats = parsed
		}
	}

	adaptiveRecentActivityStats = false
	var adaptiveValue string = queryValues.Get("adaptive_show_recent_activity_stats")

	if adaptiveValue == "true" {
		adaptiveRecentActivityStats = true
	}

	showLatestActivity = false
	var showLatestValue string = queryValues.Get("show_latest_activity")

	if showLatestValue == "true" {
		showLatestActivity = true
	}

	return username, limitCount, startTime, endTime, repositoryNameFilters, repositoryTopicFilters, includeCommits, includePullRequests, includeIssues, showRecentActivityStats, adaptiveRecentActivityStats, showLatestActivity, nil
}

// formatTimeAgo formats the duration since the given time as a human-readable string.
func (controller *ContributedRepositoriesController) formatTimeAgo(targetTime time.Time) string {
	var duration time.Duration = time.Since(targetTime)
	var hours float64 = duration.Hours()

	// Less than 1 hour: use minutes.
	if hours < 1 {
		var minutes int = int(duration.Minutes())

		if minutes <= 1 {
			return "1 minute ago"
		}

		return fmt.Sprintf("%d minutes ago", minutes)
	}

	// Less than 24 hours (1 day): use hours.
	if hours < 24 {
		var hoursInt int = int(hours)

		if hoursInt == 1 {
			return "1 hour ago"
		}

		return fmt.Sprintf("%d hours ago", hoursInt)
	}

	// Less than 168 hours (7 days / 1 week): use days.
	if hours < 168 {
		var days int = int(hours / 24)

		if days == 1 {
			return "1 day ago"
		}

		return fmt.Sprintf("%d days ago", days)
	}

	// Less than 720 hours (30 days / 1 month): use weeks.
	if hours < 720 {
		var weeks int = int(hours / 168)

		if weeks == 1 {
			return "1 week ago"
		}

		return fmt.Sprintf("%d weeks ago", weeks)
	}

	// Less than 8760 hours (365 days / 1 year): use months.
	if hours < 8760 {
		var months int = int(hours / 720)

		if months == 1 {
			return "1 month ago"
		}

		return fmt.Sprintf("%d months ago", months)
	}

	// More than 1 year: use years.
	var years int = int(hours / 8760)

	if years == 1 {
		return "1 year ago"
	}

	return fmt.Sprintf("%d years ago", years)
}

// RegisterRoutes registers the controller's endpoints to the provided ServeMux.
func (controller *ContributedRepositoriesController) RegisterRoutes(router *http.ServeMux) {
	// Explicitly map HTTP verbs and paths.
	router.HandleFunc("GET /v1/contributed-repositories", controller.HandleGetContributedRepositories)
	router.HandleFunc("GET /v1/contributed-repositories/markdown", controller.HandleGetContributedRepositoriesMarkdown)
}

// HandleGetContributedRepositories handles the request to get recently contributed repositories.
func (controller *ContributedRepositoriesController) HandleGetContributedRepositories(responseWriter http.ResponseWriter, request *http.Request) {
	// 1. Parse Request Parameters (Query String).
	var username string
	var limitCount int
	var startTime time.Time
	var endTime time.Time
	var repositoryNameFilters []string
	var repositoryTopicFilters []string
	var includeCommits bool
	var includePullRequests bool
	var includeIssues bool
	var showRecentActivityStats int
	var adaptiveRecentActivityStats bool
	var parseError error

	username, limitCount, startTime, endTime, repositoryNameFilters, repositoryTopicFilters, includeCommits, includePullRequests, includeIssues, showRecentActivityStats, adaptiveRecentActivityStats, _, parseError = controller.parseQueryParameters(request)

	if parseError != nil {
		http.Error(responseWriter, parseError.Error(), http.StatusBadRequest)
		return
	}

	// 2. Call Core Business Logic (Service Layer).
	var requestContext context.Context = request.Context()
	var results []*domain.ContributedRepository
	var executeError error

	results, executeError = controller.useCase.Execute(
		requestContext,
		username,
		startTime,
		endTime,
		limitCount,
		repositoryNameFilters,
		repositoryTopicFilters,
		includeCommits,
		includePullRequests,
		includeIssues,
		showRecentActivityStats,
		adaptiveRecentActivityStats,
	)

	if executeError != nil {
		http.Error(responseWriter, "Failed to fetch contributed repositories: "+executeError.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Map Domain Models to DTOs.
	var responseDTOs []model.ContributedRepository = make([]model.ContributedRepository, 0, len(results))
	var result *domain.ContributedRepository

	for _, result = range results {
		var responseModel model.ContributedRepository = model.ContributedRepository{
			Name:        result.Repository.FullName,
			ActiveAt:    result.ActiveAt,
			Description: result.Repository.Description,
			IsOwner:     result.IsOwner,
		}
		responseDTOs = append(responseDTOs, responseModel)
	}

	// 4. Serialize Response.
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)

	var jsonEncoder *json.Encoder = json.NewEncoder(responseWriter)
	var encodingError error = jsonEncoder.Encode(responseDTOs)

	if encodingError != nil {
		http.Error(responseWriter, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleGetContributedRepositoriesMarkdown handles the request to get recently contributed repositories in Markdown format.
func (controller *ContributedRepositoriesController) HandleGetContributedRepositoriesMarkdown(responseWriter http.ResponseWriter, request *http.Request) {
	// 1. Parse Request Parameters (Query String).
	var username string
	var limitCount int
	var startTime time.Time
	var endTime time.Time
	var repositoryNameFilters []string
	var repositoryTopicFilters []string
	var includeCommits bool
	var includePullRequests bool
	var includeIssues bool
	var showRecentActivityStats int
	var adaptiveRecentActivityStats bool
	var showLatestActivity bool
	var parseError error

	username, limitCount, startTime, endTime, repositoryNameFilters, repositoryTopicFilters, includeCommits, includePullRequests, includeIssues, showRecentActivityStats, adaptiveRecentActivityStats, showLatestActivity, parseError = controller.parseQueryParameters(request)

	if parseError != nil {
		http.Error(responseWriter, parseError.Error(), http.StatusBadRequest)
		return
	}

	// Parse title for markdown.
	var title string = request.URL.Query().Get("title")
	if title == "" {
		title = "### 📦 Project"
	}

	// 2. Call Core Business Logic (Service Layer).
	var requestContext context.Context = request.Context()
	var results []*domain.ContributedRepository
	var executeError error

	results, executeError = controller.useCase.Execute(
		requestContext,
		username,
		startTime,
		endTime,
		limitCount,
		repositoryNameFilters,
		repositoryTopicFilters,
		includeCommits,
		includePullRequests,
		includeIssues,
		showRecentActivityStats,
		adaptiveRecentActivityStats,
	)

	if executeError != nil {
		http.Error(responseWriter, "Failed to fetch contributed repositories: "+executeError.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate latest active time.
	var latestActiveAt time.Time

	if len(results) > 0 {
		// Results are already sorted by ActiveAt descending in the UseCase.
		latestActiveAt = results[0].ActiveAt
	}

	// 3. Serialize Response as Markdown.
	responseWriter.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	// Write Start Comment with Timestamp.
	fmt.Fprintf(responseWriter, "<!-- LATEST_ACTIVITY: %s -->\n", latestActiveAt.Format(time.RFC3339))

	// Write Title.
	fmt.Fprintf(responseWriter, "%s\n\n", title)

	// Write List.
	var result *domain.ContributedRepository

	for _, result = range results {
		var ownedTag string = ""
		if result.IsOwner {
			ownedTag = " `Owned`"
		}

		// - **[RepoName](RepoURL)** `Owned` — Description.
		// If description is empty, don't print the dash.
		var descriptionPart string = ""

		if result.Repository.Description != "" {
			descriptionPart = fmt.Sprintf(" — %s", result.Repository.Description)
		}

		fmt.Fprintf(responseWriter, "- **[%s](%s)**%s%s\n\n",
			result.Repository.Name,
			result.Repository.HTMLURL,
			ownedTag,
			descriptionPart,
		)

		// Render Activity Stats if present.
		if result.ActivityStats != nil {
			var stats []string

			if result.ActivityStats.CommitCount > 0 && includeCommits {
				stats = append(stats, fmt.Sprintf("%d Commits", result.ActivityStats.CommitCount))
			}
			// Note: Assuming IssueCount and PullRequestCount logic from usecase.
			// Currently usecase groups everything under PullRequestCount for PR fetcher,
			// but issue fetching isn't explicitly separated yet.
			// Using PullRequestCount here.
			if result.ActivityStats.PullRequestCount > 0 && includePullRequests {
				stats = append(stats, fmt.Sprintf("%d Pull Requests", result.ActivityStats.PullRequestCount))
			}
			if result.ActivityStats.IssueCount > 0 && includeIssues {
				stats = append(stats, fmt.Sprintf("%d Issues", result.ActivityStats.IssueCount))
			}

			if len(stats) > 0 {
				var timeLabel string = "Stats"
				var displayWindow int = showRecentActivityStats

				if result.ActivityStats.TimeWindow > 0 {
					displayWindow = result.ActivityStats.TimeWindow
				}

				switch displayWindow {
				case 24:
					timeLabel = "Day"
				case 168:
					timeLabel = "Week"
				case 720:
					timeLabel = "Month"
				case 8760:
					timeLabel = "Year"
				}

				fmt.Fprintf(responseWriter, "  📈 **Past %s:** %s\n\n", timeLabel, strings.Join(stats, " | "))
			}
		}

		// Render Latest Activity if requested and present.
		if showLatestActivity && result.LatestActivity != nil {
			var emoji string
			var activityTitle string = result.LatestActivity.Title
			var activityURL string = result.LatestActivity.URL
			var timeAgo string = controller.formatTimeAgo(result.LatestActivity.CreatedAt)

			switch result.LatestActivity.Type {
			case domain.ActivityTypeCommit:
				emoji = "📌" // Default.

				// Check for conventional commit prefix.
				var msgLower string = strings.ToLower(activityTitle)

				if strings.HasPrefix(msgLower, "feat") {
					emoji = "✨"
				} else if strings.HasPrefix(msgLower, "fix") {
					emoji = "🐛"
				} else if strings.HasPrefix(msgLower, "docs") {
					emoji = "📝"
				} else if strings.HasPrefix(msgLower, "refactor") {
					emoji = "♻️"
				} else if strings.HasPrefix(msgLower, "perf") {
					emoji = "⏫"
				} else if strings.HasPrefix(msgLower, "test") {
					emoji = "🧪"
				} else if strings.HasPrefix(msgLower, "style") {
					emoji = "💄"
				} else if strings.HasPrefix(msgLower, "chore") {
					emoji = "🧹"
				} else if strings.HasPrefix(msgLower, "ci") {
					emoji = "🤖"
				} else if strings.HasPrefix(msgLower, "revert") {
					emoji = "⏪"
				}
			case domain.ActivityTypePullRequest:
				emoji = "🔀"
			case domain.ActivityTypeIssue:
				if result.LatestActivity.IssueAction == "created" {
					emoji = "⚠️"
				} else {
					emoji = "💬"
				}
			}

			fmt.Fprintf(responseWriter, "  %s **Latest:** [%s](%s) (%s)\n\n", emoji, activityTitle, activityURL, timeAgo)
		}
	}

	// Write End Comment.
	fmt.Fprintf(responseWriter, "<!-- LATEST_ACTIVITY_END -->")
}
