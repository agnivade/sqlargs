package sqlargs

import (
	"fmt"
	"go/ast"
	"go/types"
	"strconv"

	"github.com/lfittl/pg_query_go"
	"golang.org/x/tools/go/analysis"
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
	RunDespiteErrors: true,
	FactTypes:        []analysis.Fact{new(foundFact)},
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

	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			// We filter only function calls.
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			// Now we need to find expressions like these in the source code.
			// db.Exec(`INSERT INTO <> (foo, bar) VALUES ($1, $2)`, param1, param2)

			// A CallExpr has 2 parts - Fun and Args.
			// A Fun can either be an Ident (Fun()) or a SelectorExpr (foo.Fun()).
			// Since we are looking for patterns like db.Exec, we need to filter only SelectorExpr
			// We will ignore dot imported functions.
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel == nil {
				return true
			}

			// A SelectorExpr(db.Exec) has 2 parts - X (db) and Sel (Exec/Query/QueryRow).
			// Now that we are inside the SelectorExpr, we need to verify 2 things -
			// 1. The function name is Exec, Query or QueryRow; because that is what we are interested in.
			// 2. The type of the selector is sql.DB, sql.Tx or sql.Stmt.
			// TODO: Also do the Context couterparts.
			if !isProperSelExpr(sel, pass.TypesInfo) {
				return true
			}
			// Length of args has to be minimum of 1 because we only take Exec, Query or QueryRow;
			// all of which have atleast 1 argument. But still writing a sanity check.
			if len(call.Args) == 0 {
				return true
			}

			arg0 := call.Args[0]
			if bl, ok := arg0.(*ast.BasicLit); ok {
				query, _ := strconv.Unquote(bl.Value) // errors seem to be ignored in vet checkers.
				tree, err := pg_query.Parse(query)
				if err != nil {
					pass.Reportf(call.Lparen, "Invalid query: %v\n", err)
				}
				fmt.Println(tree)
			}
			// // Now print the params of the query
			// for _, a := range call.Args {
			// 	fmt.Printf("[%#v] ", a)
			// }
			// fmt.Println()

			return true
		})
	}
	return nil, nil
}

func isProperSelExpr(sel *ast.SelectorExpr, typesInfo *types.Info) bool {
	// Only accept function calls for Exec, QueryRow and Query
	fnName := sel.Sel.Name
	if fnName != "Exec" &&
		fnName != "QueryRow" &&
		fnName != "Query" {
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
	// Only accept sql.DB, sql.Tx or sql.Stmt types.
	if ptr.Elem().String() != "database/sql.DB" &&
		ptr.Elem().String() != "database/sql.Tx" &&
		ptr.Elem().String() != "database/sql.Stmt" {
		return false
	}
	return true
}

// foundFact is a fact associated with functions that match -name.
// We use it to exercise the fact machinery in tests.
type foundFact struct{}

func (*foundFact) String() string { return "found" }
func (*foundFact) AFact()         {}
