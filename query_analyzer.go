package sqlargs

import (
	"go/ast"

	pg_query "github.com/lfittl/pg_query_go"
	nodes "github.com/lfittl/pg_query_go/nodes"
	"golang.org/x/tools/go/analysis"
)

func analyzeQuery(query string, call *ast.CallExpr, pass *analysis.Pass) {
	tree, err := pg_query.Parse(query)
	if err != nil {
		pass.Reportf(call.Lparen, "Invalid query: %v", err)
		return
	}
	// Analyze the parse tree for semantic errors.
	if len(tree.Statements) == 0 {
		return
	}
	rawStmt, ok := tree.Statements[0].(nodes.RawStmt)
	if !ok {
		return
	}
	switch stmt := rawStmt.Stmt.(type) {
	// 1. For insert statements, the no. of columns(if present) should be equal to no. of values.
	case nodes.InsertStmt:
		numCols := len(stmt.Cols.Items)
		if numCols == 0 {
			return
		}
		selStmt, ok := stmt.SelectStmt.(nodes.SelectStmt)
		if !ok {
			return
		}
		if len(selStmt.ValuesLists) == 0 {
			return
		}
		numValues := len(selStmt.ValuesLists[0])
		if numCols != numValues {
			pass.Reportf(call.Lparen, "No. of columns (%d) not equal to no. of values (%d)", numCols, numValues)
		}
	}
}
