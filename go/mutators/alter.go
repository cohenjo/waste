package mutators

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cohenjo/waste/go/config"
	wh "github.com/cohenjo/waste/go/utils"
	"github.com/github/gh-ost/go/base"
	"github.com/github/gh-ost/go/logic"
	"github.com/rs/zerolog/log"
)

type AlterTable struct {
	BaseChange
}



func (cng *AlterTable) Validate() error {
	return nil
}

func (cng *AlterTable) PostSteps() error {
	return nil
}

// RunChange - Runs the alter table
// alter table - will be processed by GH-OST
func (cng *AlterTable) RunChange() (string, error) {
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
		fmt.Printf("altering table on: (%s:%d) great \n", hostname, port)

		var msg string
		year, mo, day := time.Now().Date()
		migrationContext := cng.generateContext()

		migrationContext.InspectorConnectionConfig.Key.Hostname = hostname
		migrationContext.InspectorConnectionConfig.Key.Port = port
		migrationContext.AssumeMasterHostname = hostname
		migrationContext.DatabaseName = cng.DatabaseName
		migrationContext.OriginalTableName = cng.TableName
		migrationContext.AlterStatement = cng.SQLCmd

		migrator := logic.NewMigrator(migrationContext)

		err := migrator.Migrate()
		if err != nil {
			migrator.ExecOnFailureHook()
			log.Error().Err(err).Msgf("can't alter table table. %d-%d-%d", year, mo, day)
			msg = err.Error()
		} else {
			log.Info().Str("Action", "alter").Msg("done")
			msg = "change done"
		}

		log.Info().Str("Action", "alter").Msgf("%s", msg)
		// return msg, err

	}
	return "msg", err

}

// generateContext return a context used to later change schema in with gh-ost
func (cng *AlterTable) generateContext() *base.MigrationContext {

	migrationContext := base.NewMigrationContext()
	migrationContext.ConfigFile = ""

	migrationContext.CliUser = config.Config.DBUser
	migrationContext.CliPassword = config.Config.DBPasswd

	migrationContext.CountTableRows = false
	migrationContext.NullableUniqueKeyAllowed = false
	migrationContext.ApproveRenamedColumns = false
	migrationContext.SkipRenamedColumns = false
	migrationContext.IsTungsten = false
	migrationContext.DiscardForeignKeys = false
	migrationContext.SkipForeignKeyChecks = false
	migrationContext.TestOnReplica = false
	migrationContext.MigrateOnReplica = false
	migrationContext.OkToDropTable = false
	migrationContext.InitiallyDropOldTable = true
	migrationContext.InitiallyDropGhostTable = true
	migrationContext.DropServeSocket = true
	migrationContext.TimestampOldTable = false
	migrationContext.AssumeRBR = true
	migrationContext.SwitchToRowBinlogFormat = false
	migrationContext.AllowedRunningOnMaster = true
	migrationContext.AllowedMasterMaster = true
	migrationContext.ReplicaServerId = 99999
	migrationContext.ServeTCPPort = 0
	migrationContext.ServeSocketFile = fmt.Sprintf("/tmp/waste-%s-%s-%s.sock", cng.Cluster, cng.DatabaseName, cng.TableName)
	migrationContext.PostponeCutOverFlagFile = fmt.Sprintf("/tmp/waste-postpone-%s-%s-%s.flag", cng.Cluster, cng.DatabaseName, cng.TableName)
	niceRatio := float64(0.7)
	chunkSize := int64(1000)
	dmlBatchSize := int64(100)
	maxLagMillis := int64(1500)
	cutOverLockTimeoutSeconds := int64(3)
	migrationContext.CutOverType = base.CutOverAtomic

	migrationContext.CriticalLoadIntervalMilliseconds = int64(0)
	migrationContext.CriticalLoadHibernateSeconds = int64(0)
	migrationContext.SetHeartbeatIntervalMilliseconds(100)
	migrationContext.SetNiceRatio(niceRatio)
	migrationContext.SetChunkSize(chunkSize)
	migrationContext.SetDMLBatchSize(dmlBatchSize)
	migrationContext.SetMaxLagMillisecondsThrottleThreshold(maxLagMillis)
	migrationContext.SetThrottleQuery("")
	migrationContext.SetThrottleHTTP("")
	migrationContext.SetDefaultNumRetries(60)
	migrationContext.ApplyCredentials()
	if err := migrationContext.SetCutOverLockTimeoutSeconds(cutOverLockTimeoutSeconds); err != nil {
		log.Error().Err(err)
	}

	migrationContext.Noop = !(config.Config.Execute)
	// acceptSignals(migrationContext)

	return migrationContext
}
