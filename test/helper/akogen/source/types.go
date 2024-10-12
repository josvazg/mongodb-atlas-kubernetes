package source

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/metadata"
	"golang.org/x/tools/go/packages"
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

// TODO can we avoid passing pkgPath? can we detect from file info?
func (sg *SourceGen) DescribeType(pkgPath, typeName string) (*metadata.DataType, error) {
	conf := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedTypesSizes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(conf, pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %q: %w", pkgPath, err)
	}
	dt, err := sg.describeField(pkgs, typeName, typeName)
	if err != nil {
		return nil, fmt.Errorf("failed describe type %s: %w", typeName, err)
	}
	return &dt.DataType, err
}

func (sg *SourceGen) describeField(pkgs []*packages.Package, fieldName, typeName string) (*metadata.DataField, error) {
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(localTypeName(typeName))
		if obj == nil || !obj.Exported() {
			return nil, fmt.Errorf("not found type %s in package %q", typeName, sg.f.Name.Name)
		}
		switch to := obj.Type().Underlying().(type) {
		case *types.Basic:
			simpleField := metadata.NewSimpleField(fieldName, fieldTypeName(typeName))
			if to.Name() != obj.Name() {
				simpleField = simpleField.WithPrimitive(metadata.Type(to.Name()))
			}
			return simpleField, nil
		case *types.Struct:
			structField := metadata.NewStructField(fieldName, metadata.NewNamedType(fieldName, fieldTypeName(typeName)))
			for i := 0; i < to.NumFields(); i++ {
				field := to.Field(i)
				dataField, err := sg.describeFieldType(pkgs, field)
				if err != nil {
					return nil, fmt.Errorf("failed to describe field %s.%s: %w", typeName, field.Name(), err)
				}
				structField.Fields = append(structField.Fields, dataField)
			}
			return structField, nil
		default:
			return nil, fmt.Errorf("unsupported type %s", obj.Type())
		}
	}
	return nil, fmt.Errorf("not found type %s in package %q", typeName, sg.f.Name.Name)
}

func (sg *SourceGen) describeFieldType(pkgs []*packages.Package, field *types.Var) (*metadata.DataField, error) {
	if isPrimitiveType(field.Type()) {
		return metadata.NewSimpleField(field.Name(), fieldTypeName(field.Type().String())), nil
	}
	return sg.describeField(pkgs, field.Name(), field.Type().String())
}

func isPrimitiveType(t types.Type) bool {
	if basicType, ok := (t).(*types.Basic); ok {
		return basicType.String() == basicType.Underlying().String()
	}
	return false
}

func isPointer(typeName string) bool {
	return typeName[0] == '*'
}

func localTypeName(typeName string) string {
	parts := strings.Split(typeName, ".")
	return parts[len(parts)-1]
}

func fieldTypeName(typeName string) string {
	fn := localTypeName(typeName)
	if isPointer((typeName)) {
		fn = "*" + fn
	}
	return fn
}
