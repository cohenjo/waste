package mutators

import (
	wh "github.com/cohenjo/waste/go/utils"
	"github.com/rs/zerolog/log"
	"fmt"
	"encoding/json"
	"database/sql"
	"time"
	"github.com/cohenjo/waste/go/config"	
)

type RenameTable struct {
	BaseChange
}

func (cng *RenameTable) Validate() error {
	return nil
}

func (cng *RenameTable) PostSteps() error {
	return nil
}


// RunChange renames a table to keep it
// @todo: choose and implement cleanup policy
// @body: something will eventually need to remove these tables.
func (cng *RenameTable) RunChange() (string, error) {
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


func (cng *RenameTable) GetArtifact() string {
	return cng.Artifact
}

func (cng *RenameTable) GetCluster() string{
	return cng.Cluster
}

func (cng *RenameTable) GetDB() string {
	return cng.DatabaseName
}

func (cng *RenameTable) Immediate() bool {
	return true
}