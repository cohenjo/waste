package types

import "github.com/shurcooL/githubv4"

type ChangedFile struct {
	Sha          string
	Filename     string
	Status       string
	Additions    int
	deletions    int
	changes      int
	blob_url     string
	Raw_url      string
	Contents_url string
	patch        string
}

type ChangedFiles []ChangedFile

type PullRequestDetails struct {
	Id      githubv4.ID
	Title   githubv4.String
	Number  githubv4.Int
	Reviews struct {
		TotalCount githubv4.Int
	} `graphql:"reviews(states:APPROVED)"`
}

// ApprovedPullRequestsQuery represents all open, approved pull requests
type ApprovedPullRequestsQuery struct {
	Repository struct {
		DatabaseID   githubv4.Int
		URL          githubv4.URI
		PullRequests struct {
			Nodes    []PullRequestDetails
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage githubv4.Boolean
			}
		} `graphql:"pullRequests(first:$pullsFirst,states:OPEN)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// PullRequestsCommentMutation represents a change to pull request comments
type PullRequestsCommentMutation struct {
	AddComment struct {
		ClientMutationID githubv4.String
	} `graphql:"addComment(input: $input)"`
}
