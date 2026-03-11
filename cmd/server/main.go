package main

import (
	"fmt"
	stdlibHttp "net/http"
	"os"

	"github.com/joho/godotenv"
	"go.leoweyr.com/github-profile-postprocessor/internal/delivery/handler"
	internalHttp "go.leoweyr.com/github-profile-postprocessor/internal/delivery/http"
)

func main() {
	_ = godotenv.Load()

	// 1. Instantiate Application.
	var address string = "0.0.0.0:8080"
	var applicationEngine *internalHttp.Application = internalHttp.NewApplication(address)

	// 2. Instantiate ContributedRepositoriesController.
	var repositoryController *handler.ContributedRepositoriesController = handler.NewContributedRepositoriesController()

	// 3. Get Router.
	var router *stdlibHttp.ServeMux = applicationEngine.GetRouter()

	// 4. Register Routes.
	repositoryController.RegisterRoutes(router)

	// 5. Run Server.
	fmt.Printf("Server starting on %s...\n", address)
	var executionError error = applicationEngine.Run()

	if executionError != nil {
		fmt.Printf("FATAL: Server execution failed: %v\n", executionError)
		os.Exit(1)
	}
}
