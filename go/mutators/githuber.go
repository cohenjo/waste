package mutators

import (
	"context"
	"fmt"

	"github.com/cohenjo/waste/go/types"
	"github.com/shurcooL/githubv4"
)

// FetchRepoDescription fetches description of repo with owner and name.
func FetchRepoDescription(ctx context.Context, owner, name string) (types.ApprovedPullRequestsQuery, error) {
	var q types.ApprovedPullRequestsQuery
	// var q types.PullRequestsQuery

	variables := map[string]interface{}{
		"owner":      githubv4.String(owner),
		"name":       githubv4.String(name),
		"pullsFirst": githubv4.NewInt(3),
		// "states":     githubv4.PullRequestReviewState(githubv4.PullRequestReviewStateApproved),
	}

	err := QueryGithub(ctx, &q, variables)
	if err != nil {
		fmt.Println("Failed to query GitHub API v4:", err)
		return q, err
	}
	// printJSON(q)

	return q, err
}

// CommentPullRequest comments on a pull request of repo with owner and name.
func CommentPullRequest(ctx context.Context, message, name string, id githubv4.ID) (string, error) {
	var q types.PullRequestsCommentMutation

	var aci githubv4.AddCommentInput
	aci.Body = githubv4.String(message)
	aci.SubjectID = githubv4.ID(id)

	httpClient := GetHttpClient()

	client := githubv4.NewClient(httpClient)

	err := client.Mutate(ctx, &q, aci, nil)
	if err != nil {
		fmt.Println("Failed to query GitHub API v4:", err)
		return "problem", err
	}
	// printJSON(q)

	return "done", err
}

// FetchRepoDescription fetches description of repo with owner and name.
func FetchPullRequests(ctx context.Context, owner, name string) (types.PullRequestsQuery, error) {
	// var q types.ApprovedPullRequestsQuery
	var q types.PullRequestsQuery

	variables := map[string]interface{}{
		"owner":      githubv4.String(owner),
		"name":       githubv4.String(name),
		"pullsFirst": githubv4.NewInt(10),
		// "states":     githubv4.PullRequestReviewState(githubv4.PullRequestReviewStateApproved),
	}

	err := QueryGithub(ctx, &q, variables)
	if err != nil {
		fmt.Println("Failed to query GitHub API v4:", err)
		return q, err
	}
	// printJSON(q)

	return q, err
}
