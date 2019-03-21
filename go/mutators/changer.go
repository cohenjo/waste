package mutators

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cohenjo/waste/go/config"
	wh "github.com/cohenjo/waste/go/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

// Change represents a transformation waiting to happen
// swagger:model Change
type Change struct {
	Artifact     string
	Cluster      string
	DatabaseName string
	TableName    string
	ChangeType   string
	SQLCmd       string
}

// Result is the output of DB calls - do we need this??
type Result string

// ReadFromURL drills the content url to get the actual file content
func (c *Change) ReadFromURL(fileURL string, httpClient *http.Client) {

	resp, err := httpClient.Get(fileURL)
	if err != nil {
		// log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var objmap interface{}
	err = json.Unmarshal(body, &objmap)
	dnldURL := objmap.(map[string]interface{})["download_url"]
	resp, err = httpClient.Get(dnldURL.(string))
	if err != nil {
		// log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &c)
	if err != nil {
		// log.Fatal(err)
	}

}

// // acceptSignals registers for OS signals
// func acceptSignals(migrationContext *base.MigrationContext) {
// 	c := make(chan os.Signal, 1)

// 	signal.Notify(c, syscall.SIGHUP)
// 	go func() {
// 		for sig := range c {
// 			switch sig {
// 			case syscall.SIGHUP:
// 				log.Infof("Received SIGHUP. Reloading configuration")
// 				if err := migrationContext.ReadConfigFile(); err != nil {
// 					log.Errore(err)
// 				} else {
// 					migrationContext.MarkPointOfInterest()
// 				}
// 			}
// 		}
// 	}()
// }

// RunChange runs the change according to the change type
func (cng *Change) RunChange() (string, error) {

	var res string
	var err error
	switch cng.ChangeType {
	case "create":
		log.Info().Str("Action", "create").Msg("create new table - will be processed by CREATOR")
		res, err = cng.runTableCreate()
	case "alter":
		log.Info().Str("Action", "alter").Msg("alter table - will be processed by GH-OST")
		res, err = cng.runTableAlter()
	case "drop":
		log.Info().Str("Action", "drop").Msg("drop a table - You're likely an idiot - i'll keep it for now")
		res, err = cng.runTableRename()
	default:
		fmt.Println("You're an idiot - I'll just ignore and wait for you to go away")
	}
	return res, err
}

// EnrichChange tries to enrich the change with more details...
func (cng *Change) EnrichChange() {
	data, err := wh.GetBindings(cng.Artifact)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get bindings from  ")

	}
	m := make([]map[string]interface{}, 0)
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshel data ")
	}
	if len(m) == 0 {
		fmt.Println("didn't get any enrichmant")
	}
	if len(m) == 1 {
		if m[0]["ClusterType"].(string) == "mysql" {
			cng.Cluster = m[0]["ClusterName"].(string)
			cng.DatabaseName = m[0]["DBName"].(string)
		}
	}
}

/**
*  RunTableCreate simply runs the given create statement - no validation yet.
* TODO:
// 1. use some connection class - nicer
// 2. add validations:
*   a. table doesn't exist
*   b. validate DDL statement?
*/
func (cng *Change) runTableCreate() (string, error) {
	data, err := wh.GetMasters(cng.Cluster)
	if err != nil {
		log.Error().Err(err).Msgf("this is sad... %s", data)
		return "", err

	}
	m := make([]map[string]interface{}, 0)
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Error().Err(err).Msg("this is bad... ")
		return "", err
	}
	for _, server := range m {
		serverKey, ok := server["Key"].(map[string]interface{})
		if !ok {
			fmt.Printf("be angry")
		}
		hostname := serverKey["Hostname"]
		port := int(serverKey["Port"].(float64))
		fmt.Printf("creating table on: (%s:%d) great \n", hostname, port)
		fmt.Printf("running SQL> CREATE TABLE %s(%s) \n", cng.TableName, cng.SQLCmd)

		DBUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true&autocommit=true&charset=utf8mb4,utf8,latin1", config.Config.DBUser, config.Config.DBPasswd, hostname, port, cng.DatabaseName)
		db, err := sql.Open("mysql", DBUrl)
		defer db.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to open DB")
			continue
		}
		err = db.Ping()
		if err != nil {
			// do something here
			log.Info().Str("Action", "create").Msg("can't connect.")
			continue
		}

		var msg string
		sqlcmd := fmt.Sprintf("CREATE TABLE %s(%s)", cng.TableName, cng.SQLCmd)
		if config.Config.Execute {
			result, err := db.Exec(sqlcmd)
			if err != nil {
				// do something here
				log.Error().Err(err).Msg("can't create table.")
				msg = err.Error()
			} else {
				log.Info().Str("Action", "create").Msgf("%v", result)
				msg = "change done"

			}
		} else {
			msg = "Execute flag not set"
			err = nil
		}

		log.Info().Str("Action", "create").Msgf("%s", msg)
		// return msg, err
	}

	return "No Maters", err
}

// RunTableRename renames a table to keep it
// @todo: choose and implement cleanup policy
// @body: something will eventually need to remove these tables.
func (cng *Change) runTableRename() (string, error) {
	data, err := wh.GetMasters(cng.Cluster)
	if err != nil {
		log.Fatal().Err(err).Msgf("this is sad... %s", data)

	}
	m := make([]map[string]interface{}, 0)
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Fatal().Err(err).Msg("this is bad... ")
	}
	for _, server := range m {
		serverKey, ok := server["Key"].(map[string]interface{})
		if !ok {
			fmt.Printf("be angry")
		}
		hostname := serverKey["Hostname"]
		port := int(serverKey["Port"].(float64))
		fmt.Printf("creating table on: (%s:%d) great \n", hostname, port)

		DBUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true&autocommit=true&charset=utf8mb4,utf8,latin1", config.Config.DBUser, config.Config.DBPasswd, hostname, port, cng.DatabaseName)
		db, err := sql.Open("mysql", DBUrl)
		defer db.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to open DB")
			continue
		}
		err = db.Ping()
		if err != nil {
			// do something here
			log.Info().Str("Action", "drop").Msg("can't connect.")
			continue
		}

		var msg string
		year, mo, day := time.Now().Date()
		sqlcmd := fmt.Sprintf("ALTER TABLE %s.%s RENAME TO %s.__waste_%d_%d_%d_%s;", cng.DatabaseName, cng.TableName, cng.DatabaseName, year, mo, day, cng.TableName)
		if config.Config.Execute {
			result, err := db.Exec(sqlcmd)
			if err != nil {
				log.Error().Err(err).Msg("can't rename table.")
				msg = err.Error()
			} else {
				log.Info().Str("Action", "drop").Msgf("%v", result)
				msg = "change done"

			}
		} else {
			msg = "Execute flag not set"
			err = nil
		}
		log.Info().Str("Action", "drop").Msgf("%s", msg)
		// return msg, err

	}
	return "msg", err

}

// Validate functions validates that the change is good to go - improving our success rate.
func (cng *Change) Validate() bool {
	return true
}
