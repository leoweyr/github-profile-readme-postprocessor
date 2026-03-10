package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"go.leoweyr.com/github-profile-postprocessor/internal/model"
)

// ContributedRepositoriesController handles HTTP requests related to GitHub repositories.
type ContributedRepositoriesController struct {
	// TODO: Inject core service logic here
}

// NewContributedRepositoriesController creates a new instance of ContributedRepositoriesController.
func NewContributedRepositoriesController() *ContributedRepositoriesController {
	return &ContributedRepositoriesController{}
}

// RegisterRoutes registers the controller's endpoints to the provided ServeMux.
func (controller *ContributedRepositoriesController) RegisterRoutes(router *http.ServeMux) {
	// Explicitly map HTTP verbs and paths.
	router.HandleFunc("GET /v1/contributed-repositories", controller.HandleGetContributedRepositories)
}

// HandleGetContributedRepositories handles the request to get recently contributed repositories.
func (controller *ContributedRepositoriesController) HandleGetContributedRepositories(responseWriter http.ResponseWriter, request *http.Request) {
	// 1. Parse Request Parameters (Query String)
	// TODO: Implement rigorous parameter validation based on OpenAPI spec

	// 2. Call Core Business Logic (Service Layer)
	// TODO: Inject core service logic here

	// 3. Construct Mock Response (Stub)
	var mockContributedRepositories []model.ContributedRepository = []model.ContributedRepository{
		{
			Name:        "leoweyr/github-profile-readme-postprocessor",
			ActiveAt:    time.Now(),
			Description: "A tool to update GitHub profile readme automatically.",
			IsOwner:     true,
		},
		{
			Name:        "golang/go",
			ActiveAt:    time.Now().Add(-24 * time.Hour),
			Description: "The Go programming language.",
			IsOwner:     false,
		},
	}

	// 4. Serialize Response
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)

	var jsonEncoder *json.Encoder = json.NewEncoder(responseWriter)
	var encodingError error = jsonEncoder.Encode(mockContributedRepositories)

	if encodingError != nil {
		http.Error(responseWriter, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
