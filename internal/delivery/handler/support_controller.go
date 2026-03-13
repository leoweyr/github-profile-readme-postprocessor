package handler

import (
	"fmt"
	"net/http"
	"time"
)

// SupportController handles support-related markdown endpoints.
type SupportController struct{}

// NewSupportController creates a new instance of SupportController.
func NewSupportController() *SupportController {
	return &SupportController{}
}

// RegisterRoutes registers the support controller endpoints to the provided ServeMux.
func (controller *SupportController) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /v1/support/markdown", controller.HandleGetSupportMarkdown)
}

// HandleGetSupportMarkdown handles the request to get generation time and technical support information in Markdown format.
func (controller *SupportController) HandleGetSupportMarkdown(responseWriter http.ResponseWriter, _ *http.Request) {
	var generatedAtUtc time.Time = time.Now().UTC()
	var generatedAtString string = generatedAtUtc.Format("2006-01-02 15:04 UTC")

	responseWriter.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	fmt.Fprintf(responseWriter, "<div align=\"left\">\n")
	fmt.Fprintf(responseWriter, "  <sub>\n")
	fmt.Fprintf(responseWriter, "    ⚙️ <b>Automatically Updated Content</b> | \n")
	fmt.Fprintf(responseWriter, "    <i>Last synced: %s</i> | Power by <a href=\"https://github.com/leoweyr/github-profile-readme-postprocessor\">leoweyr/github-profile-readme-postprocessor</a>.\n", generatedAtString)
	fmt.Fprintf(responseWriter, "  </sub>\n")
	fmt.Fprintf(responseWriter, "</div>\n\n")
	fmt.Fprintf(responseWriter, "---\n")
}
