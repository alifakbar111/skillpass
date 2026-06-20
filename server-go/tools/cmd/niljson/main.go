package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"skillpass-server-go/tools/analyzer"
)

func main() {
	singlechecker.Main(niljson.Analyzer)
}
