package logic

import (
	"fmt"
	// "time"
	"github.com/google/uuid"

	// "github.com/cohenjo/waste/go/config"
	"github.com/cohenjo/waste/go/mutators"
	"github.com/cohenjo/waste/go/scheduler"
	"github.com/coreos/go-semver/semver"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/cohenjo/waste/go/utils"
	pb "github.com/cohenjo/waste/go/grpc/waste"
	"github.com/pingcap/parser/ast"
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
	changes map[string]*pb.ChangeStatus
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
		changes: make(map[string]*pb.ChangeStatus),
	}
}

// MangeChange - manages the change flow, validation, auditing, and execution.
func (cm *ChangeManager) MangeChange(change mutators.Change) (*pb.ChangeStatus,error) {

	// @todo: do we accept change sets? or 1 by 1?
	// change.EnrichChange()
	log.Info().Msg("MangeChange: got new change")

	// @todo: validate change
	err := change.Validate()
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate change")
		return nil,errors.New("Failed to validate change")
	}

	// @todo: audit change
	cngStatus,err := cm.storeChange(change)
	if err != nil {
		log.Error().Err(err).Msg("Failed to store change")
		return nil,errors.New("Failed to store change")
	}

	if change.Immediate() {
		status, err := change.RunChange()
		if err != nil {
			log.Error().Err(err).Msg("something went wrong during change run")
			return nil,err
		}
		log.Info().Msgf("finished migration, status: %s", status)
	} else {
		go change.RunChange()
	}

	// @todo: schedule change <== should we here or externally?
	

	// Cleanup renamed table
	// if change.ChangeType == "drop" {
	// 	year, mo, day := time.Now().Date()
	// 	dropThisTable := fmt.Sprintf("__waste_%d_%d_%d_%s;", year, mo, day, change.TableName)
	// 	log.Info().Msgf("Adding task to drop %s in %d days", dropThisTable, config.Config.GraceDays)
	// 	dropChange := mutators.DropChange{Cluster: change.Cluster, DatabaseName: change.DatabaseName, TableName: dropThisTable}

	// 	scheduler.WS.AddTask(config.Config.GraceDays, "drop", dropChange)
	// }
	// // Cleanup altered table
	// if change.ChangeType == "alter" {
	// 	dropThisTable := fmt.Sprintf("_gh_ost_%s_del;", change.TableName)
	// 	log.Info().Msgf("Adding task to drop %s in %d days", dropThisTable, config.Config.GraceDays)
	// 	dropChange := mutators.DropChange{Cluster: change.Cluster, DatabaseName: change.DatabaseName, TableName: dropThisTable}

	// 	scheduler.WS.AddTask(config.Config.GraceDays, "drop", dropChange)
	// }

	return cngStatus,err
}

func (cm *ChangeManager) storeChange(change mutators.Change) (*pb.ChangeStatus,error) {

	id, _ := uuid.NewUUID()
	cngStatus := &pb.ChangeStatus{
		// Change: change,
		Uuid: id.String(),
		ChangeState: pb.State_PENDING,
	}
	cm.changes[id.String()] = cngStatus

	log.Info().Msgf("storing change: %+v",change)
	var adv ArtifactDBVersion
	cm.db.Where(ArtifactDBVersion{Artifact: change.GetArtifact(),
		Cluster:      change.GetCluster(),
		DatabaseName: change.GetDB()}).Attrs("version", "1.0.0").FirstOrCreate(&adv)
	v := semver.New(adv.Version)
	v.BumpMinor()
	adv.Version = v.String()
	cm.db.Save(&adv)
	cm.db.Create(&VersionsChangeLog{Version: adv.Version, Change: change})

	return cngStatus,nil
}

func GenerateChange(in *pb.Change) (mutators.Change,error) {

	
	stmtNode,err := utils.Parse(in.Ddl)
	if err != nil {
		return 	nil,err
	}

	if stmtNode != nil {
		switch stmt := (*stmtNode).(type) {
		case *ast.CreateTableStmt:
			fmt.Printf( "CREATE: %+v \n",stmt)
			var change mutators.CreateTable
			change.ChangeType = "create"
			if stmt.Table.Name.String() != "" {
				change.TableName = stmt.Table.Name.String()	
			}
			if stmt.Table.Schema.String() != "" {
				change.DatabaseName = stmt.Table.Schema.String()
			}

			change.ASTNode = stmtNode
			change.InferFromAST()
			change.Artifact = in.Artifact
			change.SQLCmd = in.Ddl
			change.Leaders = in.Leaders
		
			return &change,nil

		case *ast.AlterTableStmt:      
			fmt.Printf( "UPDATE: %+v \n",stmt.Specs[0])
			var change mutators.AlterTable
			change.ChangeType = "alter"
			if stmt.Table.Name.String() != "" {
				change.TableName = stmt.Table.Name.String()	
			}
			if stmt.Table.Schema.String() != "" {
				change.DatabaseName = stmt.Table.Schema.String()
			}
			change.ASTNode = stmtNode
			change.InferFromAST()
			change.Artifact = in.Artifact
			change.Groups = in.Groups
		
			return &change,nil

		// case *ast.CreateIndexStmt:
		// 	fmt.Printf( "CREATE INDEX: %+v \n",stmt)
		// 	cng.ChangeType = "index"
			
		default:
			fmt.Printf("we only support alter and create table")
			return nil,fmt.Errorf("we only support alter and create table")
		}
	}

	return nil,fmt.Errorf("didn't find supported type")

}
 
func (cm *ChangeManager) GetChanges(filter *pb.Filter) []*pb.ChangeStatus {
	values := []*pb.ChangeStatus{}

	// if we have UUID filter we will use that
	if len(filter.Uuid) >0 {
		for _,uuid := range filter.Uuid {
			if val, ok := cm.changes[uuid]; ok {
				values = append(values, val)
			}
		} 
	} else {
		for _, v := range cm.changes {
			values = append(values, v)
		}
	}

	

	return values
}