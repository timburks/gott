package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// parse and dump the AST for a Go file

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, "../gott.go", nil, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the imports from the file's AST.
	for _, s := range f.Imports {
		fmt.Println(s.Path.Value)
	}

	ast.Print(fset, f)
}
