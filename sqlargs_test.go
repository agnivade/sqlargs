package sqlargs_test

import (
	"testing"

	"github.com/agnivade/sqlargs"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBasic(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, sqlargs.Analyzer, "basic") // loads testdata/src/basic
}

func TestEmbed(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, sqlargs.Analyzer, "embed")
}
