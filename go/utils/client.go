package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cohenjo/waste/go/config"
	"github.com/outbrain/golib/log"
)

var httpClient = setupHTTPClient()

func setupHTTPClient() *http.Client {
	// httpTimeout := time.Duration(config.ActiveNodeExpireSeconds) * time.Second
	timeout := 5 * time.Second
	// dialTimeout := func(network, addr string) (net.Conn, error) {
	// 	return net.DialTimeout(network, addr, httpTimeout)
	// }
	// httpTransport := &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: config.Config.MySQLShepherdSSLSkipVerify},
	// 	Dial:            dialTimeout,
	// 	ResponseHeaderTimeout: httpTimeout,
	// }
	// httpClient = &http.Client{Transport: httpTransport, Timeout: timeout}
	return &http.Client{Timeout: timeout}
}

// http://localhost:4000/api/master/ecom_local_snapshots
// http://localhost:4000/api/artifact/com.wixpress.authorization-server
// http://localhost:4000/api/artifact/com.wixpress.ecommerce.wix-ecommerce-catalog-reader-web

// test on:
// http://localhost:4000/api/artifact/com.wixpress.dbdev.pavlok-test
// http://localhost:4000/api/master/dbdev

func GetMasters(clusterName string) (response []byte, err error) {
	url := fmt.Sprintf("%s/%s", "master", clusterName)
	log.Debugf("http_client: calling url: %s", url)
	return getFromWeb(url)

}

func GetBindings(artifactName string) (response []byte, err error) {
	url := fmt.Sprintf("%s/%s", "artifact", artifactName)
	log.Debugf("http_client: calling url: %s", url)
	return getFromWeb(url)

}

func getFromWeb(endpoint string) (response []byte, err error) {
	url := fmt.Sprintf("http://%s/api/%s", config.Config.WebAddress, endpoint)
	log.Debugf("http_client: calling url: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// switch strings.ToLower(config.Config.AuthenticationMethod) {
	// case "basic", "multi":
	// 	req.SetBasicAuth(config.Config.HTTPAuthUser, config.Config.HTTPAuthPassword)
	// }
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return body, log.Errorf("HttpGetLeader: got %d status on %s", res.StatusCode, url)
	}

	return body, nil
}
