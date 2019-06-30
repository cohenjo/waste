package mutators

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pingcap/parser/format"
	"github.com/pingcap/parser/ast"
	pb "github.com/cohenjo/waste/go/grpc/waste"

	// "github.com/cohenjo/waste/go/config"
	wh "github.com/cohenjo/waste/go/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

// Change represents a transformation waiting to happen
// swagger:model Change
type BaseChange struct {
	Artifact     string `json:artifact",omitempty"`
	Cluster      string `json:cluster",omitempty"`
	DatabaseName string `json:db_name",omitempty"`
	TableName    string `json:table_name",omitempty"`
	ChangeType   string `json:change_type",omitempty"`
	SQLCmd       string `json:ddl",omitempty"`
	ASTNode *ast.StmtNode `json:",omitempty"`
	Leaders              []*pb.Host   `json:"leaders,omitempty"`
	Groups               []*pb.GHosts `json:"groups,omitempty"`
}

type Change interface {
	// Validate functions validates that the change is good to go - improving our success rate.
	Validate() error
	RunChange() (string,error)
	PostSteps() error
	GetArtifact() string
	GetCluster() string
	GetDB() string
}

// Result is the output of DB calls - do we need this??
type Result string

// ReadFromURL drills the content url to get the actual file content
func (c *BaseChange) ReadFromURL(fileURL string, httpClient *http.Client) {

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


// EnrichChange tries to enrich the change with more details...
func (cng *BaseChange) EnrichChange() {
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

func (cng *BaseChange) InferFromAST() {
	if cng.ASTNode != nil {
		switch stmt := (*cng.ASTNode).(type) {
		case *ast.CreateTableStmt:
			fmt.Printf( "CREATE: %+v \n",stmt)
			cng.ChangeType = "create"
			if stmt.Table.Name.String() != "" {
				cng.TableName = stmt.Table.Name.String()	
			}
			if stmt.Table.Schema.String() != "" {
				cng.DatabaseName = stmt.Table.Schema.String()
			}

			// _ = stmt.Restore(ctx)
		case *ast.AlterTableStmt:      
			var sb strings.Builder
			flags := format.DefaultRestoreFlags
			sb.Reset()
			ctx := format.NewRestoreCtx(flags,&sb)
			fmt.Printf( "UPDATE: %+v \n",stmt.Specs[0])
			cng.ChangeType = "alter"
			stmt.Specs[0].Restore(ctx)
			cng.SQLCmd = sb.String()
			// _ = stmt.Restore(ctx)
		case *ast.CreateIndexStmt:
			fmt.Printf( "CREATE INDEX: %+v \n",stmt)
			cng.ChangeType = "index"
			
		default:
			fmt.Printf("we only support alter and create table")
		}
	}
}


func (c *BaseChange) Validate() error {
	return nil
}
func (c *BaseChange) RunChange() (string,error){
	return "",nil
}
func (c *BaseChange) PostSteps() error {
	return nil
}
