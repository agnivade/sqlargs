package sqlargs_test

import (
	"testing"

	"github.com/agnivade/sqlargs"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestQueries(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, sqlargs.Analyzer, "a") // loads testdata/src/a/a.go.
}
