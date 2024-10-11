package source

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type SourceGen struct {
	fst *token.FileSet
	f   *ast.File
}

func New(sourceFile string) (*SourceGen, error) {
	fst := token.NewFileSet()
	f, err := parser.ParseFile(fst, sourceFile, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source file %v: %v", sourceFile, err)
	}
	return &SourceGen{fst: fst, f: f}, nil
}
