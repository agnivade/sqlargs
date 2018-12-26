// package main runs the sqlargs analyzer.
package main

import (
	"github.com/agnivade/sqlargs"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(sqlargs.Analyzer) }
