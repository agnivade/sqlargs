package sqlargs

import (
	"fmt"
	"go/ast"
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
		if f.Name != nil && f.Name.Name == "backend" {
			ast.Inspect(f, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					sel, isSel := call.Fun.(*ast.SelectorExpr)
					if isSel && sel.Sel != nil {
						switch sel.Sel.Name {
						case "Exec", "QueryRow", "Query":
							if len(call.Args) > 0 {
								arg0 := call.Args[0]
								if bl, ok := arg0.(*ast.BasicLit); ok {
									query, err := strconv.Unquote(bl.Value)
									if err != nil {
										fmt.Println(err)
										return true
									}
									_, err = pg_query.Parse(query)
									if err != nil {
										pass.Reportf(call.Lparen, "Invalid query: %v\n", err)
									}
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
				return true
			})
		}
	}
	return nil, nil
}

// foundFact is a fact associated with functions that match -name.
// We use it to exercise the fact machinery in tests.
type foundFact struct{}

func (*foundFact) String() string { return "found" }
func (*foundFact) AFact()         {}
