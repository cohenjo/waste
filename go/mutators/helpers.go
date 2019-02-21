package mutators

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"net/http"
	"time"

	"github.com/cohenjo/waste/go/config"
	"github.com/outbrain/golib/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	// "github.com/graphql-go/relay/examples/starwars"
	// "github.com/github/orchestrator/go/inst" - see this for the instance type we get back from Orch
)

type Server struct {
	HostName string
	Port     int
}

// GetArtifactServerDuo return a master/replica duo used to later change schema in with gh-ost
func GetArtifactServerDuo(serverChannel chan<- *Server, clusterID string) {
	fmt.Println("**********************************************************")

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
		// DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	var orcBaseApi string = "config.Config.OrcBaseAPI"
	var username string = config.Config.OrcUsername
	var passwd string = config.Config.OrcPasswd

	fmt.Printf("Cluster ID: %s \n", clusterID)
	url_template := "http://%s/master/mysql_%s"

	req, err := http.NewRequest("GET", fmt.Sprintf(url_template, orcBaseApi, clusterID), nil)
	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		log.Criticale(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Criticale(err)
	}
	// fmt.Printf("%s", body)
	var objmap interface{}
	err = json.Unmarshal(body, &objmap)

	// fmt.Printf("map: %s \n", objmap)
	tm := objmap.(map[string]interface{})["Key"]
	// fmt.Printf("tm: %s \n", tm)
	var masterServer = toServer(tm.(map[string]interface{}))
	fmt.Printf("Master host: %s, port: %d \n", masterServer.HostName, masterServer.Port)

	// consider using  "cluster-osc-slaves/:clusterHint" as the hosts - see

	// We have the master now let's get his replica!
	rep_templte := "http://%s/instance-replicas/%s/%d"
	rreq, err := http.NewRequest("GET", fmt.Sprintf(rep_templte, orcBaseApi, masterServer.HostName, masterServer.Port), nil)
	rreq.SetBasicAuth(username, passwd)
	resp, err = client.Do(rreq)
	if err != nil {
		log.Criticale(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &objmap)

	tm = objmap.([]interface{})[0].(map[string]interface{})["Key"]
	var repServer = toServer(tm.(map[string]interface{}))

	fmt.Printf("Replica host: %s, port: %d \n", repServer.HostName, repServer.Port)
	fmt.Println("**********************************************************")

	serverChannel <- masterServer
}

func toServer(i map[string]interface{}) *Server {
	return &Server{
		HostName: i["Hostname"].(string),
		Port:     int(i["Port"].(float64))}
}

func GetHttpClient() *http.Client {
	token := config.Config.GithubToken
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return httpClient
}

func QueryGithub(ctx context.Context, q interface{}, variables map[string]interface{}) error {

	httpClient := GetHttpClient()
	client := githubv4.NewClient(httpClient)

	err := client.Query(ctx, q, variables)
	if err != nil {
		log.Critical("Failed to query GitHub API v4:", err)
	}
	return err
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
