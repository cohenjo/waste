package mutators

import (
	"encoding/json"
	"fmt"
	"testing"

	wh "github.com/cohenjo/waste/go/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/outbrain/golib/log"
)

func TestEnrich(t *testing.T) {

	cng := &Change{Artifact: "com.org.jony-test"}
	cng.EnrichChange()

	t.Logf("###############################################################################################")
	t.Logf("after rich: %+v", cng)

}

func TestCreateTable(t *testing.T) {

	cng := &Change{Artifact: "com.wixpress.greyhound-es-testapp", DatabaseName: "greyhound_db", ChangeType: "create", TableName: "users_full_view", SQLCmd: `(
		kafka_key VARCHAR(200) NOT NULL,
		value VARCHAR(10000) NOT NULL,
		PRIMARY KEY(kafka_key)
		)`}
	cng.RunChange()

	t.Logf("###############################################################################################")
	t.Logf("after rich: %+v", cng)

}

func TestCreateLocalTable(t *testing.T) {

	cng := &Change{Artifact: "com.org.jony-test-local", ChangeType: "create", TableName: "avitalTest", SQLCmd: `(i int, v varchar(256))`}
	cng.RunChange()

	t.Logf("###############################################################################################")
	t.Logf("after rich: %+v", cng)

}

func TestGetMaster(t *testing.T) {

	cng := &Change{Artifact: "com.org.jony-test"}
	cng.EnrichChange()

	data, err := wh.GetMasters(cng.Cluster)
	if err != nil {
		log.Fatalf("this is sad... %s, %v", data, err)

	}
	m := make([]map[string]interface{}, 0)
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Fatalf("this is bad... %v", err)
	}
	for _, server := range m {
		serverKey, ok := server["Key"].(map[string]interface{})
		if !ok {
			fmt.Printf("be angry")
		}
		hostname := serverKey["Hostname"]
		port := int(serverKey["Port"].(float64))
		t.Logf("got host: %s, port: %d \n", hostname, port)

	}
	t.Logf("###############################################################################################")
	t.Logf("after rich: %+v", cng)

}
