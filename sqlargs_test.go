package sqlargs

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBasic(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "basic") // loads testdata/src/basic
}

func TestEmbed(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "embed")
}

func TestSqlx(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "sqlx")
}

func TestStripVendor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Strip",
			input: "github.com/godwhoa/upboat/vendor/github.com/jmoiron/sqlx",
			want:  "github.com/jmoiron/sqlx",
		},
		{
			name:  "Ignore",
			input: "github.com/jmoiron/sqlx",
			want:  "github.com/jmoiron/sqlx",
		},
		{
			name:  "Vendor in pkg URL",
			input: "github.com/vendor/upboat/vendor/github.com/jmoiron/sqlx",
			want:  "github.com/jmoiron/sqlx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripVendor(tt.input); got != tt.want {
				t.Errorf("stripVendor() = %v, want %v", got, tt.want)
			}
		})
	}
}
