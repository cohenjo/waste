package http

import (
	"errors"

	"github.com/hashicorp/consul/api"
)

type cluster struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Ipv4     string `json:"ipv4"`
	Port     int    `json:"port"`
}

var clusterList = []cluster{
	{ID: 1, Name: "great", Hostname: "greathost", Ipv4: "192.168.1.1", Port: 3306},
	{ID: 2, Name: "failure", Hostname: "greath2", Ipv4: "192.168.1.2", Port: 3306},
}

// Return a list of all the clusters
func getAllClusters() []cluster {
	config := api.DefaultConfig()
	config.Address = "dbmng-mysql-orchestrator0a.42.wixprod.net:8500"

	client, err := api.NewClient(config)
	if err != nil {
		panic(err)
	}

	// Get a handle to the KV API
	kv := client.KV()
	keys, _, err := kv.Keys("db/mysql/master/mysql_", "/", nil)
	if err != nil {
		panic(err)
	}
	for i, key := range keys {
		clu := cluster{ID: i, Name: key}
		clusterList = append(clusterList, clu)

	}
	return clusterList
}

func getClusterByID(id int) (*cluster, error) {
	for _, a := range clusterList {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, errors.New("Article not found")
}
