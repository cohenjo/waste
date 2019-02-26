package mutators

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	// _ "github.com/go-sql-driver/mysql"
	"github.com/cohenjo/waste/go/config"
	wh "github.com/cohenjo/waste/go/utils"
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
	cng.enrichChange()
	var res string
	var err error
	switch cng.ChangeType {
	case "create":
		log.Info().Str("Action", "create").Msg("create new table - will be processed by CREATOR")
		res, err = cng.runTableCreate()
	// case "alter":
	// 	fmt.Println("alter existing table - will be processed by GH-OST")
	// 	res, err = RunGHOstChange(Config.DBUser, Config.DBPasswd, masterHost.HostName, masterHost.Port, cng.DatabaseName, cng.TableName, cng.SQLCmd)
	case "drop":
		log.Info().Str("Action", "drop").Msg("drop a table - You're likely an idiot - i'll keep it for now")
		res, err = RunTableRename()
	default:
		fmt.Println("You're an idiot - I'll just ignore and wait for you to go away")
	}
	return res, err
}

// enrichChange tries to enrich the change with more details...
func (cng *Change) enrichChange() {
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
		fmt.Printf("running SQL> CREATE TABLE %s%s \n", cng.TableName, cng.SQLCmd)

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
		sqlcmd := fmt.Sprintf("CREATE TABLE %s%s", cng.TableName, cng.SQLCmd)
		result, err := db.Exec(sqlcmd)
		if err != nil {
			// do something here
			log.Error().Err(err).Msg("can't create table.")
			msg = err.Error()
		} else {
			log.Info().Str("Action", "create").Msgf("%v", result)
			msg = "change done"

		}
		log.Info().Str("Action", "create").Msgf("%s", msg)

	}

	return "msg", err
}

// RunTableRename renames a table to keep it
// @todo: choose and implement cleanup policy
// @body: something will eventually need to remove these tables.
func RunTableRename() (string, error) {
	return "bbb", nil

	// log.Info("user: %s, pass: %s \n", user, passwd)
	// DBUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true&autocommit=true&charset=utf8mb4,utf8,latin1", user, passwd, dbhost, port, dbname)
	// db, err := sql.Open("mysql", DBUrl)
	// defer db.Close()
	// if err != nil {
	// 	log.Fatal("failed to open DB", err)
	// }
	// err = db.Ping()
	// if err != nil {
	// 	// do something here
	// 	log.Info("can't connect.\n")
	// }

	// result, err := db.Exec("select 1 from dual")
	// if err != nil {
	// 	// do something here

	// 	log.Infof("%v", err)
	// 	log.Info("can't select dual")
	// } else {
	// 	log.Infof("%v", result)
	// }
	// var msg string
	// altercmd := fmt.Sprintf("ALTER TABLE %s RENAME TO WASTE_bck_%s;", tablename, tablename)
	// result, err = db.Exec(altercmd)
	// if err != nil {
	// 	// do something here
	// 	log.Info("can't rename table.\n")
	// 	log.Infof("%v", err)
	// 	msg = err.Error()
	// } else {
	// 	log.Infof("%v", result)
	// 	msg = "change done"
	// }

	// return msg, err
}

// GetArtifactServerDuo return a master/replica duo used to later change schema in with gh-ost
// func RunGHOstChange(user string, passwd string, dbhost string, port int, dbname string, tablename string, altercmd string) (string, error) {

// 	migrationContext := base.GetMigrationContext()

// 	migrationContext.CliUser = user
// 	migrationContext.CliPassword = passwd
// 	migrationContext.InspectorConnectionConfig.Key.Hostname = dbhost
// 	migrationContext.AssumeMasterHostname = dbhost
// 	migrationContext.InspectorConnectionConfig.Key.Port = port
// 	migrationContext.DatabaseName = dbname
// 	migrationContext.OriginalTableName = tablename
// 	migrationContext.CountTableRows = false
// 	migrationContext.NullableUniqueKeyAllowed = false
// 	migrationContext.ApproveRenamedColumns = false
// 	migrationContext.SkipRenamedColumns = false
// 	migrationContext.IsTungsten = false
// 	migrationContext.DiscardForeignKeys = false
// 	migrationContext.SkipForeignKeyChecks = false
// 	migrationContext.TestOnReplica = false
// 	migrationContext.MigrateOnReplica = false
// 	migrationContext.OkToDropTable = false
// 	migrationContext.InitiallyDropOldTable = true
// 	migrationContext.InitiallyDropGhostTable = true
// 	migrationContext.DropServeSocket = true
// 	migrationContext.TimestampOldTable = false
// 	migrationContext.AssumeRBR = true
// 	migrationContext.SwitchToRowBinlogFormat = false
// 	migrationContext.AlterStatement = altercmd
// 	migrationContext.AllowedRunningOnMaster = true
// 	migrationContext.AllowedMasterMaster = true
// 	migrationContext.ReplicaServerId = 99999
// 	migrationContext.ServeTCPPort = 0
// 	migrationContext.ServeSocketFile = ""

// 	niceRatio := float64(0)
// 	chunkSize := int64(1000)
// 	dmlBatchSize := int64(100)
// 	executeFlag := true
// 	maxLagMillis := int64(1500)
// 	cutOverLockTimeoutSeconds := int64(3)
// 	migrationContext.CutOverType = base.CutOverAtomic

// 	migrationContext.CriticalLoadIntervalMilliseconds = int64(0)
// 	migrationContext.CriticalLoadHibernateSeconds = int64(0)
// 	migrationContext.SetHeartbeatIntervalMilliseconds(100)
// 	migrationContext.SetNiceRatio(niceRatio)
// 	migrationContext.SetChunkSize(chunkSize)
// 	migrationContext.SetDMLBatchSize(dmlBatchSize)
// 	migrationContext.SetMaxLagMillisecondsThrottleThreshold(maxLagMillis)
// 	migrationContext.SetThrottleQuery("")
// 	migrationContext.SetThrottleHTTP("")
// 	migrationContext.SetDefaultNumRetries(60)
// 	migrationContext.ApplyCredentials()
// 	if err := migrationContext.SetCutOverLockTimeoutSeconds(cutOverLockTimeoutSeconds); err != nil {
// 		log.Errore(err)
// 	}

// 	if migrationContext.ServeSocketFile == "" {
// 		migrationContext.ServeSocketFile = fmt.Sprintf("/tmp/gh-ost.%s.%s.sock", migrationContext.DatabaseName, migrationContext.OriginalTableName)
// 	}

// 	migrationContext.Noop = !(executeFlag)
// 	log.Infof("starting gh-ost %+v", AppVersion)
// 	// acceptSignals(migrationContext)

// 	migrator := logic.NewMigrator()
// 	err := migrator.Migrate()
// 	if err != nil {
// 		migrator.ExecOnFailureHook()
// 		log.Fatale(err)
// 	}
// 	return "Done", err
// }
