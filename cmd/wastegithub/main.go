package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	helpers "github.com/cohenjo/waste/go/mutators"
	"github.com/cohenjo/waste/go/types"
	"github.com/outbrain/golib/log"
)

func main() {

	log.SetLevel(log.INFO)
	log.Infof("WASTE: Hello, world.\n")

	clio := helpers.CLIOptions{}
	clio.ReadArgs()
	helpers.Config = clio
	helpers.InitArtifactDetails()

	owner := helpers.Config.GithubOwner
	repoName := helpers.Config.GithubRepo

	helpers.GetQueue()
	fmt.Printf("################################################################################################################\n")

	q, err := helpers.FetchRepoDescription(context.Background(), owner, repoName)
	if err != nil {
		log.Criticale(err)
	}

	for i, pr := range q.Repository.PullRequests.Nodes {
		fmt.Printf("%d) %s is numbered: %d\n ", i, pr.Title, pr.Number)
		if pr.Reviews.TotalCount >= 1 {
			fmt.Printf("found an approved PR - do it and then mutate it!!\n")
			executePullRequest(repoName, owner, pr, clio.Execute)

			// https://developer.github.com/v3/pulls/#merge-a-pull-request-merge-button
			// not sure if I should also close using https://developer.github.com/v3/pulls/#update-a-pull-request or is it done...

		} else if pr.Reviews.TotalCount == 0 {
			log.Infof("Found PR which Needs Review: %s, %d\n", pr.Title, pr.Number)
		}
	}

}

// TODO - consider using: https://developer.github.com/v4/previews/#pull-requests-preview
func executePullRequest(name string, owner string, pr types.PullRequestDetails, execute bool) {
	// For now we'll need to use GET /repos/:owner/:repo/pulls/:number/files to get the files...

	httpClient := helpers.GetHttpClient()
	urlTemplate := "https://api.github.com/repos/%s/%s/pulls/%d/files"

	resp, err := httpClient.Get(fmt.Sprintf(urlTemplate, owner, name, pr.Number))
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
		log.Infof("%d ) %s\n", index, element.Filename)
		// fmt.Printf("%d ) %s\n", index, element)
		var cng helpers.Change
		cng.ReadFromURL(element.Contents_url, httpClient)

		log.Infof("%s\n", cng.SQLCmd)

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
		log.Infof("%v\n", cng)
		if execute {
			res, err := cng.RunChange(masterHost)
			if err != nil {
				log.Criticale(err)
			}
			log.Info("res: %s \n", res)
			helpers.CommentPullRequest(context.Background(), res, "cohenjo", pr.Id)
		} else {
			fmt.Printf("don't execute!")
		}

	}
	log.Info("Amazing! change was done - merge & close the pull request")
	// PUT /repos/:owner/:repo/pulls/:number/merge
	// merge_template := "https://api.github.com/repos/%s/%s/pulls/%d/merge"
	// var jsonStr = []byte(`{"commit_title":"you've been wasted."}`)
	// req, err := http.NewRequest(http.MethodPut, fmt.Sprintf(merge_template, owner, name, pr.Number), bytes.NewBuffer(jsonStr))
	// resp, err = httpClient.Do(req)

}
