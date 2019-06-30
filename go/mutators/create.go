package mutators

import (
	"github.com/rs/zerolog/log"
	"fmt"
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
*/
func (cng *CreateTable) RunChange() (string, error) {
	log.Info().Msgf("Running create table on: %v",cng)
	var err error
	for _, server := range cng.Leaders {

		fmt.Printf("creating table on: (%s:%d) great \n", server.Hostname, server.Port)
		fmt.Printf("running SQL> %s ;\n", cng.SQLCmd)

		DBUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true&autocommit=true&charset=utf8mb4,utf8,latin1", config.Config.DBUser, config.Config.DBPasswd, server.Hostname, server.Port, cng.DatabaseName)
		fmt.Printf("DB URL> %s ;\n", DBUrl)
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
		sqlcmd := cng.SQLCmd
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

func (cng *CreateTable) GetArtifact() string {
	return cng.Artifact
}

func (cng *CreateTable) GetCluster() string{
	return cng.Cluster
}

func (cng *CreateTable) GetDB() string {
	return cng.DatabaseName
}
