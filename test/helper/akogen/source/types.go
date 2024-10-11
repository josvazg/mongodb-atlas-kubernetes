package source

import (
	"go/ast"
	"strings"
)

func (sg *SourceGen) FindAnnotatedType(annotation string) *ast.TypeSpec {
	var found *ast.TypeSpec
	grabType := false
	ast.Inspect(sg.f, func(n ast.Node) bool {
		if comment, ok := n.(*ast.Comment); ok {
			if strings.Contains(comment.Text, annotation) {
				grabType = true
				return false
			}
		} else if typeDecl, ok := n.(*ast.TypeSpec); ok {
			if typeDecl.Doc != nil {
				comment := typeDecl.Doc.Text()
				if strings.Contains(comment, annotation) {
					found = typeDecl
					return false
				}
			} else if grabType {
				found = typeDecl
				grabType = false
				return false
			}
		}
		return true
	})
	return found
}

func (sg *SourceGen) DescribeType(typeSpec *ast.TypeSpec) {

}
