package mutators

import (
	wh "github.com/cohenjo/waste/go/utils"
	"github.com/rs/zerolog/log"
	"fmt"
	"encoding/json"
	"database/sql"
	"github.com/cohenjo/waste/go/config"
)

type CreateTable struct {
	BaseChange
}


func (cng *CreateTable) Validate() error {
	return nil
}

func (cng *CreateTable) PostSteps() error {
	return nil
}
/**
*  RunTableCreate simply runs the given create statement - no validation yet.
* TODO:
// 1. use some connection class - nicer
// 2. add validations:
*   a. table doesn't exist
*   b. validate DDL statement?
*/
func (cng *CreateTable) RunChange() (string, error) {
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