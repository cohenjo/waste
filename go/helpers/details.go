package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-chef/chef"
	"github.com/go-redis/redis"
	"github.com/outbrain/golib/log"
)

// Details is the artifact details
type Details struct {
	cluser string
	db     string
	user   string
	pass   string
}

//ArtifactDetails is a map holding datails for each artifact
type ArtifactDetails map[string]Details

var m ArtifactDetails

func InitArtifactDetails() {

	// m := make(ArtifactDetails)

	// read a client key
	key, err := ioutil.ReadFile(Config.ChefKey)
	if err != nil {
		log.Error("Couldn't read key.pem:", err)
		os.Exit(1)
	}
	log.Debugf("the KEY: %s\n", string(key))
	log.Debugf("the URL: %s\n", Config.ChefBaseURL)
	log.Debugf("the USER: %s\n", Config.ChefUser)

	// build a client
	client, err := chef.NewClient(&chef.Config{
		Name: Config.ChefUser,
		Key:  string(key),
		// goiardi is on port 4545 by default. chef-zero is 8889
		BaseURL: Config.ChefBaseURL,
	})
	if err != nil {
		log.Error("Issue setting up client:", err)
		os.Exit(1)
	}

	rclient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rclient.Ping().Result()
	log.Infof(pong, err)

	val2, err := rclient.Get("last_sync").Result()
	if err == redis.Nil {
		log.Infof("not synced")
	} else if err != nil {
		log.Criticale(err)
	} else {
		log.Infof("last_sync: %s, good enough \n", val2)
		return
	}

	// List MySQL data bag items

	dbgList, err := client.DataBags.ListItems("mysql")
	if err != nil {
		log.Critical("Issue listing bags:", err)
	}

	// Print out the list
	for j := range *dbgList {
		fmt.Printf("data for cluster: %s\n", j)

		dbi, err := client.DataBags.GetItem("mysql", j)
		// fmt.Println(dbi)
		// describe(dbi)

		// dbu := dbi.(map[string]interface{})["users"]
		dbu := drill(dbi, "users")
		if dbu != nil {
			for k := range dbu.(map[string]interface{}) {
				det := drill(dbu, k)
				// det := dbu.(map[string]interface{})[k]
				// dm, ok := det.([]interface{})
				// var dbName string
				dbName := drill(det, "db_name").(string)
				// if ok != true {
				// 	dbName = det.(map[string]interface{})["db_name"].(string)
				// } else {
				// 	dbName = dm[0].(map[string]interface{})["db_name"].(string)
				// }

				// m[k] = Details{
				// 	cluser: "test",
				// 	db:     "test db",
				// 	user:   "test user",
				// 	pass:   "test pass",
				// }

				err = rclient.HSet("artdel", k, dbName).Err()
				if err != nil {
					panic(err)
				}
				err = rclient.HSet("art_cluster", k, j).Err()
				if err != nil {
					panic(err)
				}

			}
		}
	}

	err = rclient.Set("last_sync", time.Now().Unix(), 10*time.Hour).Err()
	if err != nil {
		panic(err)
	}

}

//GetArtifactDB returns the DB setting for this artifact
func GetArtifactDB(artifact string) string {
	rclient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	val, err := rclient.HGet("artdel", artifact).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("art db:", val)
	return val
}

// GetArtifactCluster returns the cluster for this artifact
func GetArtifactCluster(idChannel chan<- string, artifact string) {
	rclient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	val, err := rclient.HGet("art_cluster", artifact).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("art cluster:", val)
	idChannel <- val
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
