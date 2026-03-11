package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
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
}

// HandleGetContributedRepositories handles the request to get recently contributed repositories.
func (controller *ContributedRepositoriesController) HandleGetContributedRepositories(responseWriter http.ResponseWriter, request *http.Request) {
	// 1. Parse Request Parameters (Query String).
	var queryValues url.Values = request.URL.Query()

	var username string = queryValues.Get("username")

	if username == "" {
		http.Error(responseWriter, "Missing required parameter: username", http.StatusBadRequest)
		return
	}

	var limitCount int = 3
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
	var endTime time.Time = time.Now()
	value = queryValues.Get("until")

	if value != "" {
		var parsed time.Time
		var parseError error
		parsed, parseError = time.Parse(time.RFC3339, value)

		if parseError == nil {
			endTime = parsed
		}
	}

	var startTime time.Time = endTime.AddDate(0, 0, -30)
	value = queryValues.Get("since")

	if value != "" {
		var parsed time.Time
		var parseError error
		parsed, parseError = time.Parse(time.RFC3339, value)

		if parseError == nil {
			startTime = parsed
		}
	}

	var repositoryNameFilter string = queryValues.Get("repository_name_contains")
	var repositoryTopicFilter string = queryValues.Get("repository_topic_contains")

	var includeCommits bool = true
	value = queryValues.Get("include_commits")

	if value == "false" {
		includeCommits = false
	}

	var includePullRequests bool = true
	value = queryValues.Get("include_pull_requests")

	if value == "false" {
		includePullRequests = false
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
		repositoryNameFilter,
		repositoryTopicFilter,
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
