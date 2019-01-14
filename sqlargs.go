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
	if !imports(pass.Pkg, "database/sql") && !imports(pass.Pkg, "github.com/jmoiron/sqlx") {
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
	// Only accept function calls for Exec, QueryRow and Query and their Context/sqlx counterparts
	fnName := sel.Sel.Name
	switch fnName {
	case "Exec", "QueryRow", "Query":
	case "ExecContext", "QueryRowContext", "QueryContext":
	case "MustExec", "QueryRowx", "Queryx":
	case "MustExecContext", "QueryRowxContext", "QueryxContext":
	default:
		return false
	}
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

	// If the object is directly *sql.DB
	if isSqlObj(nTyp.Obj(), false) {
		return true
	}

	// Otherwise, it can be a struct which embeds *sql.DB
	u := nTyp.Underlying()
	st, ok := u.(*types.Struct)
	if !ok {
		return false
	}
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if f.Embedded() && isSqlObj(f, true) { // check if the embedded field is *sql.DB-ish or not.
			return true
		}
	}
	return false
}

// isSqlObj reports whether the object is of type sql.DB-ish or not.
func isSqlObj(obj types.Object, embedded bool) bool {
	if embedded {
		if !imports(obj.Pkg(), "database/sql") {
			return false
		}
	} else {
		pkgPath := stripVendor(obj.Pkg().Path())
		if pkgPath != "database/sql" && pkgPath != "github.com/jmoiron/sqlx" {
			return false
		}
	}
	name := obj.Name()
	// Only accept sql.DB, sql.Tx or sql.Stmt types.
	if name != "DB" && name != "Tx" && name != "Stmt" {
		return false
	}
	return true
}

func imports(pkg *types.Package, path string) bool {
	for _, imp := range pkg.Imports() {
		if stripVendor(imp.Path()) == path {
			return true
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
