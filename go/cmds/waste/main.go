package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cohenjo/waste/go/helpers"
	"github.com/cohenjo/waste/go/types"
	"github.com/outbrain/golib/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	fmt.Printf("Hello, world.\n")
	log.SetLevel(log.INFO)

	clio := helpers.CLIOptions{}
	clio.ReadArgs()
	helpers.Config = clio
	helpers.InitArtifactDetails()

	owner := helpers.Config.GithubOwner
	repoName := helpers.Config.GithubRepo

	q, err := fetchRepoDescription(context.Background(), owner, repoName)
	if err != nil {
		// Handle error.
		log.Criticale(err)
	}

	token := helpers.Config.GithubToken
	// os.Getenv("GITHUB_TOKEN")

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	for i, pr := range q.Repository.PullRequests.Nodes {
		fmt.Printf("%d) %s is numbered: %d\n ", i, pr.Title, pr.Number)
		if pr.Reviews.TotalCount >= 1 {
			fmt.Printf("found an approved PR - do it and then mutate it!!\n")
			if clio.Execute {
				fmt.Printf("let's execute!")
				executePullRequest(repoName, owner, pr, httpClient)
			} else {
				fmt.Printf("don't execute!")
			}

			// https://developer.github.com/v3/pulls/#merge-a-pull-request-merge-button
			// not sure if I should also close using https://developer.github.com/v3/pulls/#update-a-pull-request or is it done...

		}
	}

}

// TODO - consider using: https://developer.github.com/v4/previews/#pull-requests-preview
func executePullRequest(name string, owner string, pr types.PullRequestDetails, httpClient *http.Client) {
	// For now we'll need to use GET /repos/:owner/:repo/pulls/:number/files to get the files...
	url_template := "https://api.github.com/repos/%s/%s/pulls/%d/files"

	resp, err := httpClient.Get(fmt.Sprintf(url_template, owner, name, pr.Number))
	if err != nil {
		log.Criticale(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var cfs types.ChangedFiles
	err = json.Unmarshal(body, &cfs)
	if err != nil {
		log.Criticale(err)
	}

	for index, element := range cfs {
		fmt.Printf("%d ) %s\n", index, element.Filename)
		// fmt.Printf("%d ) %s\n", index, element)
		var cng helpers.Change
		cng.ReadFromURL(element.Contents_url, httpClient)

		fmt.Printf("%s\n", cng.SQLCmd)

		c1 := make(chan string)
		c2 := make(chan *helpers.Server)

		// TODO: Use channels to do this in parallel
		go helpers.GetArtifactCluster(c1, cng.Artifact)
		clusterID := <-c1
		go helpers.GetArtifactServerDuo(c2, clusterID)
		masterHost := <-c2

		log.Infof("received clusterId: ", clusterID)
		log.Infof("received master host: ", masterHost)

		log.Infof("%s\n", masterHost)
		res, err := cng.RunChange(masterHost)
		if err != nil {
			log.Criticale(err)
		}
		log.Info("res: %s \n", res)
		commentPullRequest(context.Background(), res, "cohenjo", pr.Id)

	}
	log.Info("Amazing! change was done - merge & close the pull request")
	// PUT /repos/:owner/:repo/pulls/:number/merge
	// merge_template := "https://api.github.com/repos/%s/%s/pulls/%d/merge"
	// var jsonStr = []byte(`{"commit_title":"you've been wasted."}`)
	// req, err := http.NewRequest(http.MethodPut, fmt.Sprintf(merge_template, owner, name, pr.Number), bytes.NewBuffer(jsonStr))
	// resp, err = httpClient.Do(req)

}

func describe(i interface{}) {
	fmt.Printf("(%v, %T)\n", i, i)
}

func drill(i interface{}, key string) interface{} {
	switch v := i.(type) {
	case map[string]interface{}:
		return v[key]
	case []interface{}:
		return drill(v[0], key)
	default:
		return ""
	}
}

// fetchRepoDescription fetches description of repo with owner and name.
func fetchRepoDescription(ctx context.Context, owner, name string) (types.ApprovedPullRequestsQuery, error) {
	var q types.ApprovedPullRequestsQuery

	variables := map[string]interface{}{
		"owner":      githubv4.String(owner),
		"name":       githubv4.String(name),
		"pullsFirst": githubv4.NewInt(3),
	}
	token := helpers.Config.GithubToken
	// os.Getenv("GITHUB_TOKEN")

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	err := client.Query(ctx, &q, variables)
	if err != nil {
		fmt.Println("Failed to query GitHub API v4:", err)
		return q, err
	}
	// printJSON(q)

	return q, err
}

// fetchRepoDescription fetches description of repo with owner and name.
func commentPullRequest(ctx context.Context, message, name string, id githubv4.ID) (string, error) {
	var q types.PullRequestsCommentMutation

	var aci githubv4.AddCommentInput
	aci.Body = githubv4.String(message)
	aci.SubjectID = githubv4.ID(id)

	token := helpers.Config.GithubToken
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	err := client.Mutate(ctx, &q, aci, nil)
	if err != nil {
		fmt.Println("Failed to query GitHub API v4:", err)
		return "problem", err
	}
	printJSON(q)

	return "done", err
}

// printJSON prints v as JSON encoded with indent to stdout. It panics on any error.
func printJSON(v interface{}) {
	w := json.NewEncoder(os.Stdout)
	w.SetIndent("", "\t")
	err := w.Encode(v)
	if err != nil {
		panic(err)
	}
}
