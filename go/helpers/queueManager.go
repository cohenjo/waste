package helpers

import (
	"context"
	"fmt"

	"github.com/outbrain/golib/log"
)

func GetQueue() {
	owner := Config.GithubOwner
	repoName := Config.GithubRepo

	q2, err := FetchPullRequests(context.Background(), owner, repoName)
	if err != nil {
		log.Criticale(err)
	}

	for i, pr := range q2.Repository.PullRequests.Nodes {
		fmt.Printf("%d) %s is numbered: %d\n ", i, pr.Title, pr.Number)
		fmt.Printf("%d) closed: %v, merged: %v, number of reviews: %d\n ", i, pr.Closed, pr.Merged, pr.Reviews.TotalCount)
	}
	fmt.Printf("################################################################################################################\n")
}
