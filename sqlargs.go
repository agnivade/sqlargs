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

// validExprs contain all the valid selector expressions to check in the code,
// keyed by their package import path.
var validExprs = map[string]map[string]bool{
	"database/sql": {
		"DB.Exec":              true,
		"DB.ExecContext":       true,
		"DB.QueryRow":          true,
		"DB.QueryRowContext":   true,
		"DB.Query":             true,
		"DB.QueryContext":      true,
		"Tx.Exec":              true,
		"Tx.ExecContext":       true,
		"Tx.QueryRow":          true,
		"Tx.QueryRowContext":   true,
		"Tx.Query":             true,
		"Tx.QueryContext":      true,
		"Stmt.Exec":            true,
		"Stmt.ExecContext":     true,
		"Stmt.QueryRow":        true,
		"Stmt.QueryRowContext": true,
		"Stmt.Query":           true,
		"Stmt.QueryContext":    true,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Getting the list of import paths.
	var pkgs []string
	for pkg := range validExprs {
		pkgs = append(pkgs, pkg)
	}

	// We ignore packages that do not import the required paths.
	if !imports(pass.Pkg, true, pkgs...) {
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
		// XXX: This is a heuristic. Most DB code which takes context always ends with "Context"
		// and takes the ctx as the first param. But there is no guarantee for this.
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
	// Get the type info of X of the selector.
	typ, ok := typesInfo.Types[sel.X]
	if !ok {
		return false
	}

	var nTyp *types.Named
	switch t := typ.Type.(type) {
	case *types.Pointer:
		// If it is a pointer, get the element
		nTyp = t.Elem().(*types.Named)
	case *types.Named:
		nTyp = t
	}

	if nTyp == nil {
		return false
	}

	fnName := sel.Sel.Name
	objName := nTyp.Obj().Name()

	// Check valid selector expressions for a match.
	for path, obj := range validExprs {
		// If the object is a direct match.
		if imports(nTyp.Obj().Pkg(), false, path) && obj[objName+"."+fnName] {
			return true
		}

		// Otherwise, it can be a struct which embeds *sql.DB
		u := nTyp.Underlying()
		st, ok := u.(*types.Struct)
		if !ok {
			continue
		}
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			// check if the embedded field is *sql.DB-ish or not.
			if f.Embedded() && imports(f.Pkg(), true, path) && obj[f.Name()+"."+fnName] {
				return true
			}
		}
	}
	return false
}

func imports(pkg *types.Package, checkImports bool, paths ...string) bool {
	if pkg == nil {
		return false
	}
	if checkImports {
		for _, imp := range pkg.Imports() {
			for _, p := range paths {
				if imp.Path() == p {
					return true
				}
			}
		}
	} else {
		for _, p := range paths {
			if pkg.Path() == p {
				return true
			}
		}
	}
	return false
}

// stripVendor strips out the vendor path prefix
func stripVendor(pkgPath string) string {
	idx := strings.LastIndex(pkgPath, "vendor/")
	if idx < 0 {
		return pkgPath
	}
	// len("vendor/") == 7
	return pkgPath[idx+7:]
}
