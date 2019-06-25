package utils



import (
	"fmt"
	"strings"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/format"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

// This example show how to parse a text sql into ast.
func Parse(query string) (*ast.StmtNode,error) {

	// 0. make sure import parser_driver implemented by TiDB(user also can implement own driver by self).
	// and add `import _ "github.com/pingcap/tidb/types/parser_driver"` in the head of file.

	// 1. Create a parser. The parser is NOT goroutine safe and should
	// not be shared among multiple goroutines. However, parser is also
	// heavy, so each goroutine should reuse its own local instance if
	// possible.
	p := parser.New()

	// 2. Parse a text SQL into AST([]ast.StmtNode).
	// src := "CREATE TABLE foo (a SMALLINT UNSIGNED, b INT UNSIGNED) "
	// src := "ALTER TABLE foo ADD COLUMN bar VARCHAR(256)"
	// src := "select * from tbl where id = 1"
	stmtNode,  err := p.ParseOneStmt(query, "", "")
	if err != nil {
		fmt.Printf("ERROR: %v",err)
		return nil,err
	}

	var sb strings.Builder
	flags := format.DefaultRestoreFlags
	sb.Reset()
	ctx := format.NewRestoreCtx(flags,&sb)

	// 3. Use AST to do cool things.
	if stmtNode != nil {
		switch stmt := stmtNode.(type) {
		case *ast.CreateTableStmt:
			fmt.Printf( "CREATE: %+v \n",stmt)
			_ = stmt.Restore(ctx)
		case *ast.AlterTableStmt:      
			fmt.Printf( "UPDATE: %+v \n",stmt.Specs[0])
			stmt.Specs[0].Restore(ctx)
			// _ = stmt.Restore(ctx)
		case *ast.CreateIndexStmt:
			fmt.Printf( "CREATE INDEX: %+v \n",stmt)
			_ = stmt.Restore(ctx)
		default:
			fmt.Printf("we only support alter and create table")
		}
	}
	fmt.Printf("SB: %s",sb.String())
	return &stmtNode,nil
}

// {
// 	db: <name>,
// 	artifact: <>
// 	changes: [
// 		{}
// 		"add column"
// 		"drop index"
// 	]
// }
