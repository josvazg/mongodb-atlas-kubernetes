package source

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/metadata"
)

func (sg *SourceGen) FindAnnotatedType(annotation string) *ast.TypeSpec {
	for _, files := range sg.pkgs {
		for _, f := range files {
			found := findAnnotatedType(f, annotation)
			if found != nil {
				return found
			}
		}
	}
	return nil
}

func findAnnotatedType(f *ast.File, annotation string) *ast.TypeSpec {
	var found *ast.TypeSpec
	grabType := false
	ast.Inspect(f, func(n ast.Node) bool {
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

func (sg *SourceGen) DescribeType(typeName string) (*metadata.DataType, error) {
	pkgPath, err := PkgPath(sg.srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect package path at %q: %w", sg.srcPath, err)
	}
	conf := pkgsConfig()
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

func (sg *SourceGen) DescribeInterface(interfaceName string) (*metadata.Interface, error) {
	pkgPath, err := PkgPath(sg.srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect package path at %q: %w", sg.srcPath, err)
	}
	conf := pkgsConfig()
	pkgs, err := packages.Load(conf, pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %q: %w", pkgPath, err)
	}
	di := &metadata.Interface{Name: interfaceName}
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(localTypeName(interfaceName))
		if obj == nil || !obj.Exported() {
			continue
		}
		interfaceType, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			return nil, fmt.Errorf("%s is not an interface but a %T", interfaceName, obj.Type().Underlying())
		}
		for i := 0; i < interfaceType.NumMethods(); i++ {
			fn := interfaceType.Method(i)
			di.Operations = append(di.Operations, describeFunc(fn))
		}
		return di, nil
	}
	return nil, fmt.Errorf("not found type %s in package %q", interfaceName, sg.pkgName)
}

func describeFunc(fn *types.Func) metadata.FunctionSignature {
	fs := metadata.FunctionSignature{Name: fn.Name()}
	signature := fn.Signature()
	if signature == nil {
		return fs
	}
	params := signature.Params()
	paramsCount := 0
	if params != nil {
		paramsCount = params.Len()
	}
	for i := 0; i < paramsCount; i++ {
		arg := params.At(i)
		fs.Args = append(fs.Args, metadata.NewNamedType(arg.Name(), fieldTypeName(arg.Type().String())))
	}
	returns := signature.Results()
	returnsCount := 0
	if returns != nil {
		returnsCount = returns.Len()
	}
	for i := 0; i < returnsCount; i++ {
		ret := returns.At(i)
		fs.Returns = append(fs.Returns, metadata.NewNamedType("", fieldTypeName(ret.Type().String())))
	}
	return fs
}

func pkgsConfig() *packages.Config {
	return &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedTypesSizes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
	}
}

func (sg *SourceGen) describeField(pkgs []*packages.Package, fieldName, typeName string) (*metadata.DataField, error) {
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(localTypeName(typeName))
		if obj == nil || !obj.Exported() {
			continue
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
	return nil, fmt.Errorf("not found type %s in package %q", typeName, sg.pkgName)
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
	if !strings.Contains(typeName, "/") {
		return typeName
	}
	parts := strings.Split(typeName, ".")
	return parts[len(parts)-1]
}

func fieldTypeName(typeName string) string {
	fn := localTypeName(typeName)
	if isPointer(typeName) {
		fn = "*" + fn
	}
	return fn
}
