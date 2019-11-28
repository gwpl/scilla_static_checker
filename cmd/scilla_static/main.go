package main

import (
	"fmt"
	"gitlab.chainsecurity.com/ChainSecurity/common/scilla_static/pkg/ast"
	"gitlab.chainsecurity.com/ChainSecurity/common/scilla_static/pkg/ir"
	"gitlab.chainsecurity.com/ChainSecurity/common/scilla_static/pkg/souffle"
	"io/ioutil"
	"os"
	"path"
)

func main() {
	jsonPath := os.Args[1]
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	cm, err := ast.Parse_mod(byteValue)
	if err != nil {
		panic(err)
	}

	b := ir.BuildCFG(cm)

	analysisFolder := "./souffle_analysis"
	factsOutFolder := path.Join(analysisFolder, "facts_out")
	factsInFolder := path.Join(analysisFolder, "facts_in")

	err = souffle.MakeCleanFolder(factsOutFolder)
	if err != nil {
		panic(err)
	}

	err = souffle.MakeCleanFolder(factsInFolder)
	if err != nil {
		panic(err)
	}

	ir.DumpFacts(b, factsInFolder)

	souffle.RunSouffle("souffle_analysis/analysis.dl", "souffle_analysis/facts_in", "souffle_analysis/facts_out")
	result := souffle.ReadOutput("souffle_analysis/facts_out/vulnerability.csv")
	fmt.Println("======RESULTS======")
	fmt.Println(result)
}
