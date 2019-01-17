package sqlargs

import (
	"go/ast"

	pg_query "github.com/lfittl/pg_query_go"
	nodes "github.com/lfittl/pg_query_go/nodes"
	"golang.org/x/tools/go/analysis"
)

func analyzeQuery(query string, call *ast.CallExpr, pass *analysis.Pass, checkArgs bool) {
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
		if !checkArgs {
			return
		}
		numParams := numParams(selStmt.ValuesLists[0])
		args := len(call.Args[1:])
		// A safe check is to just check if args are less than no. of params. If this is true,
		// then there has to be an error somewhere. On the contrary, if there are less params
		// found than args, then it just means we haven't parsed the query well enough and there are
		// other parts of the query which use the other arguments.
		if args < numParams {
			pass.Reportf(call.Lparen, "No. of args (%d) is less than no. of params (%d)", args, numParams)
		}
	}
}

//numParams returns the count of unique paramters.
func numParams(params []nodes.Node) int {
	num := 0
	// posMap is used to keep track of unique positional parameters.
	posMap := make(map[int]bool)
	for _, p := range params {
		switch t := p.(type) {
		case nodes.ParamRef:
			if !posMap[t.Number] {
				num++
				posMap[t.Number] = true
			}
		case nodes.TypeCast:
			if pRef, ok := t.Arg.(nodes.ParamRef); ok {
				if !posMap[pRef.Number] {
					num++
					posMap[pRef.Number] = true
				}
			}
		}
	}
	return num
}
