package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-chef/chef"
)

func main() {
	// read a client key
	key, err := ioutil.ReadFile("/Users/jony/.chef/jonyc.pem")
	if err != nil {
		fmt.Println("Couldn't read key.pem:", err)
		os.Exit(1)
	}

	// build a client
	client, err := chef.NewClient(&chef.Config{
		Name: "jonyc",
		Key:  string(key),
		// goiardi is on port 4545 by default. chef-zero is 8889
		BaseURL: "http://chef.wixpress.com",
		SkipSSL: false,
	})
	if err != nil {
		fmt.Println("Issue setting up client:", err)
	}

	databags, err := client.DataBags.List()
	if err != nil {
		fmt.Errorf("DataBags.List returned error: %v", err)
	}
	// Print out the list
	fmt.Println(databags)
}
