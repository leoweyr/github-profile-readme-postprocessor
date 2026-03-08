package fetcher

import (
	"context"
	"fmt"
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

	var httpClient = oauth2.NewClient(context.Background(), tokenSource)
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
	var err error
	createdIssues, err = fetcher.fetchCreatedIssues(context, username, startTime, endTime)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch created issues: %w", err)
	}

	allActivities = append(allActivities, createdIssues...)

	// 2. Fetch commented issues (commenter:username).
	var commentedIssues []*domain.Issue
	commentedIssues, err = fetcher.fetchCommentedIssues(context, username, startTime, endTime)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch commented issues: %w", err)
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
		var err error

		result, response, err = fetcher.client.Search.Issues(context, query, searchOptions)

		if err != nil {
			return nil, err
		}

		for _, item := range result.Issues {
			var repositoryName string = extractRepositoryName(item.GetHTMLURL())

			var activity *domain.Issue = &domain.Issue{
				Title:          item.GetTitle(),
				HTMLURL:        item.GetHTMLURL(),
				RepositoryName: repositoryName,
				CreatedAt:      item.GetCreatedAt().Time,
				Number:         item.GetNumber(),
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
		var err error

		result, response, err = fetcher.client.Search.Issues(context, query, searchOptions)
		if err != nil {
			return nil, err
		}

		for _, item := range result.Issues {
			var repositoryName string = extractRepositoryName(item.GetHTMLURL())
			if repositoryName == "" {
				continue
			}

			// Parse owner and repo from repository name (owner/repo).
			var parts []string = strings.Split(repositoryName, "/")

			if len(parts) != 2 {
				continue
			}

			var owner string = parts[0]
			var repo string = parts[1]

			// Fetch comments for this issue to find the ones by the user in the time range.
			var comments []*domain.Issue
			comments, err = fetcher.fetchIssueComments(context, owner, repo, item.GetNumber(), username, startTime, endTime, item)

			if err != nil {
				// Log error but continue with other issues.
				// Print to stdout for debugging since no logger is available.
				fmt.Printf("Warning: failed to fetch comments for %s/%s#%d: %v\n", owner, repo, item.GetNumber(), err)
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

func (fetcher *IssueActivityFetcher) fetchIssueComments(context context.Context, owner, repo string, issueNumber int, username string, startTime, endTime time.Time, issue *github.Issue) ([]*domain.Issue, error) {
	var activities []*domain.Issue
	var listOptions *github.IssueListCommentsOptions = &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Since:       &startTime, // Optimization: only fetch comments updated since startTime.
	}

	for {
		var comments []*github.IssueComment
		var response *github.Response
		var err error

		comments, response, err = fetcher.client.Issues.ListComments(context, owner, repo, issueNumber, listOptions)
		if err != nil {
			return nil, err
		}

		for _, comment := range comments {
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
				RepositoryName: fmt.Sprintf("%s/%s", owner, repo),
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
