package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"

	"go.leoweyr.com/github-profile-postprocessor/internal/domain"
	"go.leoweyr.com/github-profile-postprocessor/internal/domain/enums"
)

// IssueActivityFetcher handles the retrieval of issue-related activities (creation and comments).
type IssueActivityFetcher struct {
	client *github.Client
}

// NewIssueActivityFetcher creates a new instance of IssueActivityFetcher with the provided token.
func NewIssueActivityFetcher(token string) *IssueActivityFetcher {
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	var httpClient *http.Client = oauth2.NewClient(context.Background(), tokenSource)
	var client *github.Client = github.NewClient(httpClient)

	return &IssueActivityFetcher{
		client: client,
	}
}

// FetchIssueActivities retrieves issue activities (created issues and comments) for the user within the time range.
func (fetcher *IssueActivityFetcher) FetchIssueActivities(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Issue, error) {
	var allActivities []*domain.Issue

	// 1. Fetch created issues (author:username).
	var createdIssues []*domain.Issue
	var fetchError error
	createdIssues, fetchError = fetcher.fetchCreatedIssues(context, username, startTime, endTime)

	if fetchError != nil {
		return nil, fmt.Errorf("failed to fetch created issues: %w", fetchError)
	}

	allActivities = append(allActivities, createdIssues...)

	// 2. Fetch commented issues (commenter:username).
	var commentedIssues []*domain.Issue
	commentedIssues, fetchError = fetcher.fetchCommentedIssues(context, username, startTime, endTime)

	if fetchError != nil {
		return nil, fmt.Errorf("failed to fetch commented issues: %w", fetchError)
	}

	allActivities = append(allActivities, commentedIssues...)

	return allActivities, nil
}

func (fetcher *IssueActivityFetcher) fetchCreatedIssues(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Issue, error) {
	var activities []*domain.Issue
	var searchOptions *github.SearchOptions = &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "created",
		Order:       "desc",
	}

	var dateRange string = fmt.Sprintf("%s..%s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	var query string = fmt.Sprintf("author:%s type:issue created:%s", username, dateRange)

	for {
		var result *github.IssuesSearchResult
		var response *github.Response
		var searchError error

		result, response, searchError = fetcher.client.Search.Issues(context, query, searchOptions)

		if searchError != nil {
			return nil, searchError
		}

		var issue *github.Issue

		for _, issue = range result.Issues {
			var repositoryName string = extractRepositoryName(issue.GetHTMLURL())

			var activity *domain.Issue = &domain.Issue{
				Title:          issue.GetTitle(),
				HTMLURL:        issue.GetHTMLURL(),
				RepositoryName: repositoryName,
				CreatedAt:      issue.GetCreatedAt().Time,
				Number:         issue.GetNumber(),
				Action:         enums.IssueActionCreated,
			}

			activities = append(activities, activity)
		}

		if response.NextPage == 0 {
			break
		}

		searchOptions.Page = response.NextPage
	}

	return activities, nil
}

func (fetcher *IssueActivityFetcher) fetchCommentedIssues(context context.Context, username string, startTime, endTime time.Time) ([]*domain.Issue, error) {
	var activities []*domain.Issue
	var searchOptions *github.SearchOptions = &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Sort:        "updated",
		Order:       "desc",
	}

	// Search for issues updated after startTime because comments could be recent.
	// Note: The updated:>=startTime filter captures issues that might have old creation dates but new comments.
	var query string = fmt.Sprintf("commenter:%s type:issue updated:>=%s", username, startTime.Format("2006-01-02"))

	for {
		var result *github.IssuesSearchResult
		var response *github.Response
		var searchError error

		result, response, searchError = fetcher.client.Search.Issues(context, query, searchOptions)
		if searchError != nil {
			return nil, searchError
		}

		var issue *github.Issue

		for _, issue = range result.Issues {
			var repositoryName string = extractRepositoryName(issue.GetHTMLURL())
			if repositoryName == "" {
				continue
			}

			// Parse owner and repo from repository name (owner/repo).
			var parts []string = strings.Split(repositoryName, "/")

			if len(parts) != 2 {
				continue
			}

			var owner string = parts[0]
			var repoName string = parts[1]

			// Fetch comments for this issue to find the ones by the user in the time range.
			var comments []*domain.Issue
			var fetchCommentsError error
			comments, fetchCommentsError = fetcher.fetchIssueComments(context, owner, repoName, issue.GetNumber(), username, startTime, endTime, issue)

			if fetchCommentsError != nil {
				// Log error but continue with other issues.
				// Print to stdout for debugging since no logger is available.
				fmt.Printf("Warning: failed to fetch comments for %s/%s#%d: %v\n", owner, repoName, issue.GetNumber(), fetchCommentsError)
				continue
			}

			activities = append(activities, comments...)
		}

		if response.NextPage == 0 {
			break
		}

		searchOptions.Page = response.NextPage
	}

	return activities, nil
}

func (fetcher *IssueActivityFetcher) fetchIssueComments(context context.Context, owner, repositoryName string, issueNumber int, username string, startTime, endTime time.Time, issue *github.Issue) ([]*domain.Issue, error) {
	var activities []*domain.Issue
	var listOptions *github.IssueListCommentsOptions = &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Since:       &startTime, // Optimization: only fetch comments updated since startTime.
	}

	for {
		var comments []*github.IssueComment
		var response *github.Response
		var listError error

		comments, response, listError = fetcher.client.Issues.ListComments(context, owner, repositoryName, issueNumber, listOptions)
		if listError != nil {
			return nil, listError
		}

		var comment *github.IssueComment

		for _, comment = range comments {
			// Check if comment is by the user.
			if comment.GetUser().GetLogin() != username {
				continue
			}

			var createdAt time.Time = comment.GetCreatedAt().Time

			// Check time range strictly.
			if createdAt.Before(startTime) || createdAt.After(endTime) {
				continue
			}

			var activity *domain.Issue = &domain.Issue{
				Title:          issue.GetTitle(),
				HTMLURL:        comment.GetHTMLURL(),
				RepositoryName: fmt.Sprintf("%s/%s", owner, repositoryName),
				CreatedAt:      createdAt,
				Number:         issueNumber,
				Action:         enums.IssueActionCommented,
			}

			activities = append(activities, activity)
		}

		if response.NextPage == 0 {
			break
		}

		listOptions.Page = response.NextPage
	}

	return activities, nil
}

func extractRepositoryName(htmlURL string) string {
	if htmlURL == "" {
		return ""
	}

	var parts []string = strings.Split(htmlURL, "/")

	if len(parts) >= 5 {
		return fmt.Sprintf("%s/%s", parts[3], parts[4])
	}

	return ""
}
