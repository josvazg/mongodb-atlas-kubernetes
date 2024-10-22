package source

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

type SourceGen struct {
	fst     *token.FileSet
	pkgName string
	pkgs    map[string][]*ast.File
	srcPath string
}

func NewFromFile(sourceFile string) (*SourceGen, error) {
	dir := filepath.Dir(sourceFile)
	base := filepath.Base(sourceFile)
	filter := func(fInfo fs.FileInfo) bool {
		return fInfo.Name() == base
	}
	return newFromDir(dir, filter)
}

func NewFromPackage(pkgPath string) (*SourceGen, error) {
	dir, err := FilePath(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find sources for package path %q: %w", pkgPath, err)
	}
	return newFromDir(dir, nil)
}

func newFromDir(dir string, filter func(fs.FileInfo) bool) (*SourceGen, error) {
	fst := token.NewFileSet()
	pkgs, err := parser.ParseDir(fst, dir, filter, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source dir %v: %w", dir, err)
	}
	filesPerPkg := map[string][]*ast.File{}
	pkgName := ""
	for name, pkg := range pkgs {
		fileList := make([]*ast.File, 0, len(pkg.Files))
		if pkgName == "" && !strings.HasSuffix(name, "_test") {
			pkgName = name
		}
		for _, f := range pkg.Files {
			fileList = append(fileList, f)
		}
		filesPerPkg[name] = fileList
	}
	return &SourceGen{fst: fst, pkgName: pkgName, pkgs: filesPerPkg, srcPath: dir}, nil
}
