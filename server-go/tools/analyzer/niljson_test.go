package niljson_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"skillpass-server-go/tools/analyzer"
)

func TestNilJSON(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, niljson.Analyzer, "p")
}
