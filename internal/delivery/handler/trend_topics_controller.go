package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go.leoweyr.com/github-profile-postprocessor/internal/usecase"
)

// TrendTopicsController handles requests for trend topics.
type TrendTopicsController struct {
	trendTopicsUseCase *usecase.TrendTopicsUseCase
}

// NewTrendTopicsController creates a new instance of TrendTopicsController.
func NewTrendTopicsController(trendTopicsUseCase *usecase.TrendTopicsUseCase) *TrendTopicsController {
	return &TrendTopicsController{
		trendTopicsUseCase: trendTopicsUseCase,
	}
}

// GetTrendTopicsMarkdown handles the request to get trend topics in markdown format.
func (controller *TrendTopicsController) GetTrendTopicsMarkdown(responseWriter http.ResponseWriter, request *http.Request) {
	var queryValues url.Values = request.URL.Query()

	var username string = queryValues.Get("username")

	if username == "" {
		http.Error(responseWriter, "Missing required parameter: username", http.StatusBadRequest)

		return
	}

	var timeLimitHours int = 24
	var limitStr string = queryValues.Get("time_limit_hours")

	if limitStr != "" {
		var val int
		var err error
		val, err = strconv.Atoi(limitStr)

		if err == nil {
			timeLimitHours = val
		}
	}

	var highlightThresholdPercentage float64 = 0.1
	var thresholdStr string = queryValues.Get("highlight_threshold_percentage")

	if thresholdStr != "" {
		// Handle potential percentage input (e.g. "10" vs "0.1")?
		// For now assume decimal 0.1.
		var val float64
		var err error
		val, err = strconv.ParseFloat(thresholdStr, 64)

		if err == nil {
			highlightThresholdPercentage = val
		}
	}

	var limitCount int = 10
	limitStr = queryValues.Get("limit_count")

	if limitStr != "" {
		var val int
		var err error
		val, err = strconv.Atoi(limitStr)

		if err == nil {
			limitCount = val
		}
	}

	var includePrivate bool = false
	var includePrivateStr string = queryValues.Get("include_private")

	if includePrivateStr != "" {
		var val bool
		var err error
		val, err = strconv.ParseBool(includePrivateStr)

		if err == nil {
			includePrivate = val
		}
	}

	var markdown string
	var err error

	markdown, err = controller.trendTopicsUseCase.Execute(
		request.Context(),
		username,
		timeLimitHours,
		highlightThresholdPercentage,
		limitCount,
		includePrivate,
	)

	if err != nil {
		http.Error(responseWriter, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	// Set Content-Type to text/markdown (or text/plain as it is a string).
	responseWriter.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write([]byte(markdown))
}
