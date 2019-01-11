package sqlargs

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check sql query strings for correctness

The sqlargs analyser checks the parameters passed to sql queries
and the actual number of parameters written in the query string
and reports any mismatches.

This is a common occurence when updating a sql query to add/remove
a column.`

var Analyzer = &analysis.Analyzer{
	Name:             "sqlargs",
	Doc:              Doc,
	Run:              run,
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	RunDespiteErrors: true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// We ignore packages that do not import database/sql.
	hasImport := false
	for _, imp := range pass.Pkg.Imports() {
		if imp.Path() == "database/sql" {
			hasImport = true
			break
		}
	}
	if !hasImport {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	// We filter only function calls.
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		// Now we need to find expressions like these in the source code.
		// db.Exec(`INSERT INTO <> (foo, bar) VALUES ($1, $2)`, param1, param2)

		// A CallExpr has 2 parts - Fun and Args.
		// A Fun can either be an Ident (Fun()) or a SelectorExpr (foo.Fun()).
		// Since we are looking for patterns like db.Exec, we need to filter only SelectorExpr
		// We will ignore dot imported functions.
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		// A SelectorExpr(db.Exec) has 2 parts - X (db) and Sel (Exec/Query/QueryRow).
		// Now that we are inside the SelectorExpr, we need to verify 2 things -
		// 1. The function name is Exec, Query or QueryRow; because that is what we are interested in.
		// 2. The type of the selector is sql.DB, sql.Tx or sql.Stmt.
		if !isProperSelExpr(sel, pass.TypesInfo) {
			return
		}
		// Length of args has to be minimum of 1 because we only take Exec, Query or QueryRow;
		// all of which have atleast 1 argument. But still writing a sanity check.
		if len(call.Args) == 0 {
			return
		}

		// Check if it is a Context call, then re-slice the first item which is a context.
		if strings.HasSuffix(sel.Sel.Name, "Context") {
			call.Args = call.Args[1:]
		}

		arg0 := call.Args[0]
		typ, ok := pass.TypesInfo.Types[arg0]
		if !ok || typ.Value == nil {
			return
		}
		analyzeQuery(constant.StringVal(typ.Value), call, pass)
	})

	return nil, nil
}

func isProperSelExpr(sel *ast.SelectorExpr, typesInfo *types.Info) bool {
	// Only accept function calls for Exec, QueryRow and Query and their Context counterparts
	fnName := sel.Sel.Name
	if fnName != "Exec" && fnName != "ExecContext" &&
		fnName != "QueryRow" && fnName != "QueryRowContext" &&
		fnName != "Query" && fnName != "QueryContext" {
		return false
	}
	// Get the type info of X of the selector.
	typ, ok := typesInfo.Types[sel.X]
	if !ok {
		return false
	}
	ptr, ok := typ.Type.(*types.Pointer)
	if !ok {
		return false
	}
	n := ptr.Elem().(*types.Named)
	if n.Obj().Pkg().Path() != "database/sql" {
		return false
	}
	name := n.Obj().Name()
	// Only accept sql.DB, sql.Tx or sql.Stmt types.
	if name != "DB" && name != "Tx" && name != "Stmt" {
		return false
	}
	return true
}
