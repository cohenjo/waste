package logic

import (
	"fmt"
	"time"

	"github.com/cohenjo/waste/go/config"
	"github.com/cohenjo/waste/go/mutators"
	"github.com/cohenjo/waste/go/scheduler"
	"github.com/coreos/go-semver/semver"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type VersionsChangeLog struct {
	gorm.Model
	mutators.Change
	Version string
}

type ArtifactDBVersion struct {
	Artifact     string `gorm:"PRIMARY_KEY"`
	Cluster      string `gorm:"primary_key"`
	DatabaseName string `gorm:"primary_key"`
	Version      string
}

type ChangeManager struct {
	db *gorm.DB
}

var CM *ChangeManager

func SetupChangeManager() *ChangeManager {
	db, err := gorm.Open("sqlite3", "waste.db")
	if err != nil {
		panic("failed to connect database")
	}
	// defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&VersionsChangeLog{})
	db.AutoMigrate(&ArtifactDBVersion{})

	//Register tasks with the scheduler
	scheduler.WS.RegisterTask("drop", mutators.DropTask)

	return &ChangeManager{
		db: db,
	}
}

// MangeChange - manages the change flow, validation, auditing, and execution.
func (cm *ChangeManager) MangeChange(change mutators.Change) error {

	// @todo: do we accept change sets? or 1 by 1?
	change.EnrichChange()

	// @todo: validate change
	ok := change.Validate()
	if !ok {
		log.Error().Msg("Failed to validate change")
		return errors.New("Failed to validate change")
	}

	// @todo: audit change
	cm.storeChange(change)

	// @todo: schedule change <== should we here or externally?
	status, err := change.RunChange()
	if err != nil {
		log.Error().Err(err).Msg("something went wrong during change run")
		return err
	}

	// Cleanup renamed table
	if change.ChangeType == "drop" {
		year, mo, day := time.Now().Date()
		dropThisTable := fmt.Sprintf("__waste_%d_%d_%d_%s;", year, mo, day, change.TableName)
		log.Info().Msgf("Adding task to drop %s in %d days", dropThisTable, config.Config.GraceDays)
		dropChange := mutators.DropChange{Cluster: change.Cluster, DatabaseName: change.DatabaseName, TableName: dropThisTable}

		scheduler.WS.AddTask(config.Config.GraceDays, "drop", dropChange)
	}
	// Cleanup altered table
	if change.ChangeType == "alter" {
		dropThisTable := fmt.Sprintf("_gh_ost_%s_del;", change.TableName)
		log.Info().Msgf("Adding task to drop %s in %d days", dropThisTable, config.Config.GraceDays)
		dropChange := mutators.DropChange{Cluster: change.Cluster, DatabaseName: change.DatabaseName, TableName: dropThisTable}

		scheduler.WS.AddTask(config.Config.GraceDays, "drop", dropChange)
	}

	log.Info().Msgf("finished migration, status: %s", status)

	return nil
}

func (cm *ChangeManager) storeChange(change mutators.Change) {

	var adv ArtifactDBVersion
	cm.db.Where(ArtifactDBVersion{Artifact: change.Artifact,
		Cluster:      change.Cluster,
		DatabaseName: change.DatabaseName}).Attrs("version", "1.0.0").FirstOrCreate(&adv)
	v := semver.New(adv.Version)
	v.BumpMinor()
	adv.Version = v.String()
	cm.db.Save(&adv)
	cm.db.Create(&VersionsChangeLog{Version: adv.Version, Change: change})
}
