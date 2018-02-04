package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	// "github.com/graphql-go/relay/examples/starwars"
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

	var orcBaseApi string = Config.OrcBaseAPI
	var username string = Config.OrcUsername
	var passwd string = Config.OrcPasswd

	fmt.Printf("Cluster ID: %s \n", clusterID)
	url_template := "http://%s/master/mysql_%s"

	req, err := http.NewRequest("GET", fmt.Sprintf(url_template, orcBaseApi, clusterID), nil)
	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%s", body)
	var objmap interface{}
	err = json.Unmarshal(body, &objmap)

	// fmt.Printf("map: %s \n", objmap)
	tm := objmap.(map[string]interface{})["MasterKey"]
	// fmt.Printf("tm: %s \n", tm)
	var masterServer = toServer(tm.(map[string]interface{}))
	fmt.Printf("Master host: %s, port: %d \n", masterServer.HostName, masterServer.Port)

	// We have the master now let's get his replica!
	rep_templte := "http://%s/instance-replicas/%s/%d"
	rreq, err := http.NewRequest("GET", fmt.Sprintf(rep_templte, orcBaseApi, masterServer.HostName, masterServer.Port), nil)
	rreq.SetBasicAuth(username, passwd)
	resp, err = client.Do(rreq)
	if err != nil {
		log.Fatal(err)
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
