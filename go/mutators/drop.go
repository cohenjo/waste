package mutators

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cohenjo/waste/go/config"
	wh "github.com/cohenjo/waste/go/utils"
	"github.com/rs/zerolog/log"
)

type DropChange struct {
	BaseChange
}

func (cng *DropChange) Validate() error {
	return nil
}

func (cng *DropChange) PostSteps() error {
	return nil
}

func (cng *DropChange) RunChange() (string, error) {
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
		hostname := serverKey["Hostname"].(string)
		port := int(serverKey["Port"].(float64))
		msg, _ := cng.runDropTable(hostname, port)
		log.Info().Str("Action", "drop").Msgf("%s", msg)

	}
	return "msg", err
}

// runDropTable drops a table to keep it
// @todo: choose and implement cleanup policy
// @body: something will eventually need to remove these tables.
func (cng *DropChange) runDropTable(hostname string, port int) (string, error) {

	log.Info().Str("Action", "drop").Msgf("Dropping table  on: (%s:%d) ", hostname, port)
	DBUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true&autocommit=true&charset=utf8mb4,utf8,latin1", config.Config.DBUser, config.Config.DBPasswd, hostname, port, cng.DatabaseName)
	db, err := sql.Open("mysql", DBUrl)
	defer db.Close()
	if err != nil {
		log.Error().Err(err).Msg("failed to open DB")
		return "failed to open DB", err

	}
	err = db.Ping()
	if err != nil {
		// do something here
		log.Info().Str("Action", "drop").Msg("can't connect.")
		return "can't connect.", err
	}

	msg := "table dropped succesfully"
	sqlcmd := fmt.Sprintf("DROP TABLE %s.%s;", cng.DatabaseName, cng.TableName)
	_, err = db.Exec(sqlcmd)
	if err != nil {
		log.Error().Err(err).Msg("can't rename table.")
		msg = err.Error()
	}
	log.Info().Str("Action", "drop").Msgf("%s", msg)
	return msg, err

}

func DropTask(cng DropChange) {
	// var c DropChange
	// mapstructure.Decode(cng, &c)

	msg, err := cng.RunChange()
	if err != nil {
		log.Error().Err(err).Msgf("Drop change failed - this may leave trash: %s", cng.TableName)
	}
	log.Info().Msgf("Table Dropped: %s", msg)
}


func (cng *DropChange) GetArtifact() string {
	return cng.Artifact
}

func (cng *DropChange) GetCluster() string{
	return cng.Cluster
}

func (cng *DropChange) GetDB() string {
	return cng.DatabaseName
}
