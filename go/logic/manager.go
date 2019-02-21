package logic


import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/wix-system/shepherd/pkg/instances"
)



func GetCluster(clusterName string) (response []Instance, err error) {
	url := fmt.Sprintf("%s/mysql_%s", "cluster", clusterName)
	log.Debugf("http_client: calling url: %s", url)
	err = GetFromShepherd(url, &response)
	if err != nil {
		return nil, err
	}
	return response, nil

}

func GetFromShepherd(endpoint string, dest interface{}) error {
	url := fmt.Sprintf("http://%s/api/%s", config.Config.OrchestratorAddr, endpoint)
	log.Debugf("http_client: calling url: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	// switch strings.ToLower(config.Config.AuthenticationMethod) {
	// case "basic", "multi":
	// 	req.SetBasicAuth(config.Config.HTTPAuthUser, config.Config.HTTPAuthPassword)
	// }
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(dest)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return log.Errorf("HttpGetLeader: got %d status on %s", res.StatusCode, url)
	}

	return nil
}
