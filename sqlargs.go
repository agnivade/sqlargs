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
	for _, f := range pass.Files {
		// We ignore files which do not import database/sql
		hasImport := false
		for _, i := range f.Imports {
			if i.Path != nil {
				path, _ := strconv.Unquote(i.Path.Value)
				if path == "database/sql" {
					hasImport = true
					break
				}
			}
		}
		if !hasImport {
			continue
		}

		ast.Inspect(f, func(n ast.Node) bool {
			// We filter only function calls.
			if call, ok := n.(*ast.CallExpr); ok {
				// Now we need to find expressions like these in the source code.
				// db.Exec(`INSERT INTO <> (foo, bar) VALUES ($1, $2)`, param1, param2)

				// A CallExpr has 2 parts - Fun and Args.
				// A Fun can either be an Ident (Fun()) or a SelectorExpr (foo.Fun()).
				// Since we are looking for patterns like db.Exec, we need to filter only SelectorExpr
				// We will ignore dot imported functions.
				sel, isSel := call.Fun.(*ast.SelectorExpr)
				if isSel && sel.Sel != nil {
					// A SelectorExpr has 2 parts - X (db) and Sel (Exec/Query/QueryRow).
					// Now that we are inside the SelectorExpr, we need to verify 2 things -
					// 1. The function name is either Exec or Query or QueryRow; because that is what we are interested in.
					// 2. The type of the selector is either sql.DB or sql.Tx.
					switch sel.Sel.Name {
					case "Exec", "QueryRow", "Query":
						if len(call.Args) > 0 {
							if typ, ok := pass.TypesInfo.Types[sel.X]; ok {
								pt, ok := (typ.Type).(*types.Pointer)
								if ok {
									if pt.Elem().String() == "database/sql.DB" || pt.Elem().String() == "database/sql.Tx" {
										fmt.Println("match!")
									} else {
										return true
									}
								} else {
									return true
								}
							}
							// sel.X can be ident or selector Expr
							switch expr := sel.X.(type) {
							case *ast.Ident:
								fmt.Printf("%#v\n", expr.Obj)
							case *ast.SelectorExpr:
								fmt.Printf("%#v\n", expr.X)
							}

							arg0 := call.Args[0]
							if bl, ok := arg0.(*ast.BasicLit); ok {
								query, _ := strconv.Unquote(bl.Value)
								_, err = pg_query.Parse(query)
								if err != nil {
									pass.Reportf(call.Lparen, "Invalid query: %v\n", err)
								}
							}
							// // Now print the params of the query
							// for _, a := range call.Args {
							// 	fmt.Printf("[%#v] ", a)
							// }
							// fmt.Println()
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

// foundFact is a fact associated with functions that match -name.
// We use it to exercise the fact machinery in tests.
type foundFact struct{}

func (*foundFact) String() string { return "found" }
func (*foundFact) AFact()         {}
