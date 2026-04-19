package main

import (
	"fmt"
	stdlibHttp "net/http"
	"os"

	"github.com/joho/godotenv"
	"go.leoweyr.com/github-profile-postprocessor/internal/delivery/handler"
	internalHttp "go.leoweyr.com/github-profile-postprocessor/internal/delivery/http"
	"go.leoweyr.com/github-profile-postprocessor/internal/gateway/fetcher"
	"go.leoweyr.com/github-profile-postprocessor/internal/usecase"
)

func main() {
	_ = godotenv.Load()

	// 1. Configuration.
	var address string = "0.0.0.0:8080"

	var envPort string = os.Getenv("APP_LISTEN_PORT")

	if envPort != "" {
		address = "0.0.0.0:" + envPort
	}

	var githubToken string = os.Getenv("APP_GITHUB_TOKEN")

	if githubToken == "" {
		fmt.Printf("FATAL: GITHUB_TOKEN environment variable is required.\n")
		os.Exit(1)
	}

	// 2. Instantiate Dependencies (Gateways).
	var commitFetcher *fetcher.CommitFetcher = fetcher.NewCommitFetcher(githubToken)
	var issueFetcher *fetcher.IssueActivityFetcher = fetcher.NewIssueActivityFetcher(githubToken)
	var prFetcher *fetcher.PullRequestFetcher = fetcher.NewPullRequestFetcher(githubToken)
	var repoFetcher *fetcher.RepositoryFetcher = fetcher.NewRepositoryFetcher(githubToken)

	// 3. Instantiate Use Cases.
	var contributedReposUseCase *usecase.ContributedRepositoriesUseCase = usecase.NewContributedRepositoriesUseCase(
		commitFetcher,
		issueFetcher,
		prFetcher,
		repoFetcher,
	)

	var trendTopicsUseCase *usecase.TrendTopicsUseCase = usecase.NewTrendTopicsUseCase(
		commitFetcher,
		issueFetcher,
		prFetcher,
		repoFetcher,
	)

	// 4. Instantiate Application Engine.
	var applicationEngine *internalHttp.Application = internalHttp.NewApplication(address)

	// 5. Instantiate Controllers.
	var supportController *handler.SupportController = handler.NewSupportController()
	var repositoryController *handler.ContributedRepositoriesController = handler.NewContributedRepositoriesController(contributedReposUseCase)
	var trendTopicsController *handler.TrendTopicsController = handler.NewTrendTopicsController(trendTopicsUseCase)

	// 6. Get Router.
	var router *stdlibHttp.ServeMux = applicationEngine.GetRouter()

	// 7. Register Routes.
	supportController.RegisterRoutes(router)
	repositoryController.RegisterRoutes(router)
	trendTopicsController.RegisterRoutes(router)

	// 8. Run Server.
	var executionError error = applicationEngine.Run()

	if executionError != nil {
		fmt.Printf("FATAL: Server execution failed: %v\n", executionError)
		os.Exit(1)
	}
}
