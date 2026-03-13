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

// RegisterRoutes registers the controller's endpoints to the provided ServeMux.
func (controller *ContributedRepositoriesController) RegisterRoutes(router *http.ServeMux) {
	// Explicitly map HTTP verbs and paths.
	router.HandleFunc("GET /v1/contributed-repositories", controller.HandleGetContributedRepositories)
	router.HandleFunc("GET /v1/contributed-repositories/markdown", controller.HandleGetContributedRepositoriesMarkdown)
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
	err error,
) {
	var queryValues url.Values = request.URL.Query()

	username = queryValues.Get("username")

	if username == "" {
		return "", 0, time.Time{}, time.Time{}, nil, nil, false, false, fmt.Errorf("missing required parameter: username")
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

		var s string

		for _, s = range splits {
			var trimmed string = strings.Trim(s, "\"")
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

		var s string

		for _, s = range splits {
			var trimmed string = strings.Trim(s, "\"")
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

	return username, limitCount, startTime, endTime, repositoryNameFilters, repositoryTopicFilters, includeCommits, includePullRequests, nil
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
	var parseError error

	username, limitCount, startTime, endTime, repositoryNameFilters, repositoryTopicFilters, includeCommits, includePullRequests, parseError = controller.parseQueryParameters(request)

	if parseError != nil {
		http.Error(responseWriter, parseError.Error(), http.StatusBadRequest)
		return
	}

	// 2. Call Core Business Logic (Service Layer).
	var context context.Context = request.Context()
	var results []*domain.ContributedRepository
	var executeError error

	results, executeError = controller.useCase.Execute(
		context,
		username,
		startTime,
		endTime,
		limitCount,
		repositoryNameFilters,
		repositoryTopicFilters,
		includeCommits,
		includePullRequests,
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
	var parseError error

	username, limitCount, startTime, endTime, repositoryNameFilters, repositoryTopicFilters, includeCommits, includePullRequests, parseError = controller.parseQueryParameters(request)

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
	var context context.Context = request.Context()
	var results []*domain.ContributedRepository
	var executeError error

	results, executeError = controller.useCase.Execute(
		context,
		username,
		startTime,
		endTime,
		limitCount,
		repositoryNameFilters,
		repositoryTopicFilters,
		includeCommits,
		includePullRequests,
	)

	if executeError != nil {
		http.Error(responseWriter, "Failed to fetch contributed repositories: "+executeError.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Serialize Response as Markdown.
	responseWriter.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

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
		fmt.Fprintf(responseWriter, "- **[%s](%s)**%s — %s\n\n",
			result.Repository.Name,
			result.Repository.HTMLURL,
			ownedTag,
			result.Repository.Description,
		)
	}
}
