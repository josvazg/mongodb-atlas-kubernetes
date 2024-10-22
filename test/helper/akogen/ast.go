package akogen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/metadata"
)

const (
	varKey     = "var"
	typeKey    = "type"
	pathKey    = "path"
	pointerKey = "pointer"
)

var primitiveTypeNames = []string{
	"string",
	"bool",
	"int", "int8", "int16", "int32", "int64",
	"unit", "unit16", "uint32", "uint64", "uint64", "uintptr",
	"float32", "float64",
	"complex64", "complex128",
}

func NewTranslationLayerFromSourceFile(src string) (*TranslationLayer, error) {
	tlp := translatorLayerParser{parsedTypes: make(map[metadata.Type]*metadata.DataType)}

	err := tlp.parseFile(src)
	if err != nil {
		return nil, fmt.Errorf("failed to extract code generation info: %w", err)
	}

	err = tlp.parseLibrary()
	if err != nil {
		return nil, fmt.Errorf("failed to parse the external library info: %w", err)
	}

	return tlp.translationLayer(), nil
}

type translatorLayerParser struct {
	packageName      string
	externalAlias    string
	externalType     string
	externalAPIAlias string
	externalAPIType  string
	internalAlias    string
	internalPointer  bool
	parsedTypes      map[metadata.Type]*metadata.DataType
	currentType      metadata.Type
	wp               WrappedType
	err              error
}

func (tlp *translatorLayerParser) parseFile(src string) error {
	fst := token.NewFileSet()
	f, err := parser.ParseFile(fst, src, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse source file: %w", err)
	}

	pkg, err := packageFor(fst, f, filepath.Dir(src))
	if err != nil {
		return fmt.Errorf("failed to type check the package: %w", err)
	}
	tlp.packageName = pkg.Name()

	// top level comments are not walked otherwise
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			tlp.parseComment(c)
		}
	}
	ast.Walk(tlp, f)
	if tlp.err != nil {
		return fmt.Errorf("failed to parse the source types: %w", tlp.err)
	}
	expandComplexFields(tlp.wp.Internal, tlp.parsedTypes)
	if err := tlp.setExternal(); err != nil {
		return fmt.Errorf("failed to compute external type: %w", err)
	}
	if err := tlp.setExternalAPI(); err != nil {
		return fmt.Errorf("failed to compute external API: %w", err)
	}
	return nil
}

func (tlp *translatorLayerParser) parseLibrary() error {
	// libDir, err := getPackageDir(tlp.wp.Lib.Path)
	// if err != nil {
	// 	return fmt.Errorf("failed to translate the library to a path: %w", err)
	// }

	// fst := token.NewFileSet()
	// pkgs, err := parser.ParseDir(fst, libDir, nil, parser.ParseComments)
	// if err != nil {
	// 	return fmt.Errorf("failed to parse library sources: %w", err)
	// }
	// cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedSyntax}
	// pkgs, err := packages.Load(cfg, libDir)
	// if err != nil {
	// 	return fmt.Errorf("")
	// }

	// tlp.parsedTypes = make(map[Type]*DataType)
	// tlp.currentType = ""
	// tlp.err = nil
	// ast.Walk(tlp, pkg)
	return nil
}

func (tlp *translatorLayerParser) Visit(node ast.Node) ast.Visitor {
	if tlp.visit(node) {
		return tlp
	}
	return nil
}

func (tlp *translatorLayerParser) parseComment(c *ast.Comment) error {
	if strings.Contains(c.Text, "+akogen:") {
		annotation, err := parseAnnotationValue(c.Text)
		if err != nil {
			return fmt.Errorf("failed to extract annotation value: %w", err)
		}
		return tlp.parseAnnotation(annotation)
	}
	return nil
}

func (tlp *translatorLayerParser) parseAnnotation(annotation string) error {
	switch {
	case strings.Contains(annotation, "ExternalSystem:"):
		value, err := parseAnnotationValue(annotation)
		if err != nil {
			return fmt.Errorf("failed to parse ExternalSystem annotation: %w", err)
		}
		tlp.wp.ExternalName = value
	case strings.Contains(annotation, "ExternalPackage:"):
		values, err := parseAnnotationCSVValue(annotation)
		if err != nil {
			return fmt.Errorf("failed to parse ExternalPackage annotation: %w", err)
		}
		tlp.wp.Lib.Alias = values[varKey]
		tlp.wp.Lib.Path = values[pathKey]
	case strings.Contains(annotation, "ExternalType:"):
		values, err := parseAnnotationCSVValue(annotation)
		if err != nil {
			return fmt.Errorf("failed to parse ExternalType annotation: %w", err)
		}
		tlp.externalAlias = values[varKey]
		tlp.externalType = values[typeKey]
	case strings.Contains(annotation, "ExternalAPI:"):
		values, err := parseAnnotationCSVValue(annotation)
		if err != nil {
			return fmt.Errorf("failed to parse ExternalAPI annotation: %w", err)
		}
		tlp.externalAPIAlias = values[varKey]
		tlp.externalAPIType = values[typeKey]
	case strings.Contains(annotation, "WrapperType:"):
		values, err := parseAnnotationCSVValue(annotation)
		if err != nil {
			return fmt.Errorf("failed to parse WrapperType annotation: %w", err)
		}
		tlp.wp.Wrapper = metadata.NewNamedType(values[varKey], values[typeKey])
	case strings.Contains(annotation, "InternalType:"):
		values, err := parseAnnotationCSVValue(annotation)
		if err != nil {
			return fmt.Errorf("failed to parse InternalType annotation")
		}
		var ok bool
		tlp.internalAlias, ok = values[varKey]
		if !ok {
			return fmt.Errorf("missing var for InternalType")
		}
		tlp.internalPointer = isTrue(values[pointerKey])
	default:
		return fmt.Errorf("unsupported annotation %q", annotation)
	}
	return nil
}

func (tlp *translatorLayerParser) translationLayer() *TranslationLayer {
	return &TranslationLayer{
		PackageName: tlp.packageName,
		WrappedType: &tlp.wp,
	}
}

func (tlp *translatorLayerParser) visit(node ast.Node) bool {
	if tlp.err != nil || node == nil {
		return false
	}

	switch obj := node.(type) {
	case *ast.Comment:
		err := tlp.parseComment(obj)
		return tlp.visitResult(err)
	case *ast.TypeSpec:
		tlp.parseType(obj)
		return true
	case *ast.StructType:
		err := tlp.parseStructType(obj)
		return tlp.visitResult(err)
	case *ast.Ident:
		tlp.parseIdent(obj)
		return true
	default:
		return true
	}
}

func (tlp *translatorLayerParser) visitResult(err error) bool {
	if err != nil {
		tlp.err = err
		return false
	}
	return true
}

func (tlp *translatorLayerParser) parseType(ts *ast.TypeSpec) {
	typeName := ts.Name.Name
	if tlp.internalAlias != "" && tlp.wp.Internal == nil {
		if tlp.internalPointer {
			typeName = "*" + typeName
		}
		tlp.wp.Internal = metadata.NewStruct(metadata.NewNamedType(tlp.internalAlias, typeName))
	} else {
		tlp.parsedTypes[metadata.Type(typeName)] = metadata.NewSimpleDataType(metadata.UnknownName, typeName)
	}
	tlp.currentType = metadata.Type(typeName)
}

func (tlp *translatorLayerParser) parseStructType(st *ast.StructType) error {
	if st.Fields == nil {
		return nil
	}
	for _, f := range st.Fields.List {
		if len(f.Names) < 1 {
			return fmt.Errorf("failed to parse field of type %v, expected at least one name", f.Type)
		}
		name := f.Names[0].String()
		typeName := parseTypeExpr(f.Type)
		simpleField := metadata.NewSimpleField(name, typeName)
		var dataType *metadata.DataType
		if tlp.wp.Internal != nil && tlp.currentType == tlp.wp.Internal.Type {
			dataType = tlp.wp.Internal
		} else {
			dataType = tlp.parsedTypes[tlp.currentType]
		}
		dataType.Kind = metadata.Struct
		dataType.Fields = append(dataType.Fields, simpleField)
	}
	tlp.currentType = ""
	return nil
}

func parseTypeExpr(te ast.Expr) string {
	switch expr := te.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", parseTypeExpr(expr.X))
	default:
		return fmt.Sprintf("{unsupported expression %v of type %T}", te, te)
	}
}

func (tlp *translatorLayerParser) parseIdent(id *ast.Ident) {
	if id == nil || tlp.currentType == "" || id.Name == string(tlp.currentType) {
		return
	}
	dt := tlp.parsedTypes[tlp.currentType]
	if isPrimitiveTypeName(id.Name) {
		dt.WithPrimitive(metadata.Type(id.Name))
	}
}

func expandComplexFields(dt *metadata.DataType, types map[metadata.Type]*metadata.DataType) {
	for i, field := range dt.Fields {
		if isPrimitiveTypeName(string(field.Type)) {
			continue
		}
		key := field.Type.Dereference()
		fullDecl, ok := types[key]
		if ok {
			newField := metadata.NewFieldFromData(field.FieldName, fullDecl)
			newField.Type = field.Type
			dt.Fields[i] = newField
			if fullDecl.Kind != metadata.SimpleField {
				expandComplexFields(fullDecl, types)
			}
		}
	}
}

func (tlp *translatorLayerParser) setExternal() error {
	if tlp.externalType == "" {
		return fmt.Errorf("missing type value for ExternalType annotation")
	}
	externalNamedType, err := parseNamedType(tlp.externalAlias, tlp.externalType, tlp.wp.Lib.Path)
	if err != nil {
		return fmt.Errorf("failed to parse ExternalType named type: %w", err)
	}
	tlp.wp.External = metadata.NewStruct(externalNamedType)
	return nil
}

func (tlp *translatorLayerParser) setExternalAPI() error {
	if tlp.externalAPIType == "" {
		return fmt.Errorf("missing type value for ExternalAPI annotation")
	}
	externalAPINamedType, err := parseNamedType(tlp.externalAPIAlias, tlp.externalAPIType, tlp.wp.Lib.Path)
	if err != nil {
		return fmt.Errorf("failed to parse ExternalAPI named type: %w", err)
	}
	tlp.wp.ExternalAPI = externalAPINamedType
	return nil
}

func packageFor(fst *token.FileSet, f *ast.File, path string) (*types.Package, error) {
	// Create a type configuration and checker
	conf := types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
	}

	return conf.Check(path, fst, []*ast.File{f}, info)
}

func parseAssignment(s string) (string, string, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("failed to extract assignment value from %q", s)
	}
	return parts[0], strings.Trim(parts[1], " \""), nil
}

func parseAnnotationValue(s string) (string, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to extract annotation value from %s", s)
	}
	return parts[1], nil
}

// func parseAnnotationKey(s string) (string, error) {
// 	parts := strings.SplitN(s, ":", 2)
// 	if len(parts) != 2 {
// 		return "", fmt.Errorf("failed to extract annotation key from %s", s)
// 	}
// 	return parts[0], nil
// }

func parseCSVMap(value string) (map[string]string, error) {
	values := map[string]string{}
	tuples := strings.Split(value, ",")
	for _, tuple := range tuples {
		k, v, err := parseAssignment(tuple)
		if err != nil {
			return nil, fmt.Errorf("failed to parse annotation CSV values: %w", err)
		}
		values[k] = v
	}
	return values, nil
}

func parseAnnotationCSVValue(s string) (map[string]string, error) {
	value, err := parseAnnotationValue(s)
	if err != nil {
		return nil, err
	}
	return parseCSVMap(value)
}

func parseNamedType(alias, namespacedTypeName, pkgPath string) (metadata.NamedType, error) {
	dereferenced, isPtr := dereference(namespacedTypeName)
	_, typeName, err := splitNamespacedType(dereferenced)
	if err != nil {
		return metadata.NamedType{}, fmt.Errorf("failed to parse ExternalType named type: %w", err)
	}
	fullTypeName := fmt.Sprintf("%s.%s", pkgPath, typeName)
	if isPtr {
		fullTypeName = fmt.Sprintf("*%s", fullTypeName)
	}
	return metadata.NewNamedType(alias, fullTypeName), nil
}

func dereference(namespacedTypeName string) (string, bool) {
	if strings.HasPrefix(namespacedTypeName, "*") {
		return namespacedTypeName[1:], true
	}
	return namespacedTypeName, false
}

func splitNamespacedType(namespacedTypeName string) (string, string, error) {
	parts := strings.SplitN(namespacedTypeName, ".", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("failed to separate namespaced namespace.Typename like from %s", namespacedTypeName)
	}
	return parts[0], parts[1], nil
}

func isPrimitiveTypeName(expr string) bool {
	for _, primitiveTypeName := range primitiveTypeNames {
		if strings.Contains(expr, primitiveTypeName) {
			return true
		}
	}
	return false
}

func isTrue(s string) bool {
	return strings.TrimSpace(strings.ToLower(s)) == "true"
}

// func getPackageDir(packageName string) (string, error) {
// 	cmd := exec.Command("go", "list", "-mod=mod", "-f", "{{.Dir}}", packageName)
// 	out, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to query go list package %q: %w", packageName, err)
// 	}
// 	return strings.TrimSpace(string(out)), nil
// }
