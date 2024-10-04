package akogen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
)

const (
	varKey     = "var"
	typeKey    = "type"
	pathKey    = "path"
	pointerKey = "pointer"
)

func NewTranslationLayerFromSourceFile(src string) (*TranslationLayer, error) {
	fst := token.NewFileSet()
	f, err := parser.ParseFile(fst, src, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source file: %w", err)
	}

	pkg, err := packageFor(fst, f, filepath.Dir(src))
	if err != nil {
		return nil, fmt.Errorf("failed to type check the package: %w", err)
	}

	tlp := translatorLayerParser{packageName: pkg.Name()}

	err = tlp.parseFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to extract code generation info: %w", err)
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
	wp               WrappedType
	err              error
}

func (tlp *translatorLayerParser) parseFile(f *ast.File) error {
	// top level comments are not walked otherwise
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			tlp.parseComment(c)
		}
	}
	ast.Walk(tlp, f)
	return tlp.err
}

func (tlp *translatorLayerParser) Visit(node ast.Node) ast.Visitor {
	if tlp.visit(node) {
		return tlp
	}
	return nil
}

func (tlp *translatorLayerParser) parseComment(c *ast.Comment) error {
	text := c.Text
	switch {
	case strings.Contains(text, "+akogen:ExternalSystem:"):
		value, err := parseAnnotationValue(text)
		if err != nil {
			return fmt.Errorf("failed to parse ExternalSystem annotation: %w", err)
		}
		tlp.wp.ExternalName = value
	case strings.Contains(text, "+akogen:ExternalPackage:"):
		values, err := parseAnnotationCSVValue(text)
		if err != nil {
			return fmt.Errorf("failed to parse ExternalPackage annotation: %w", err)
		}
		tlp.wp.Lib.Alias = values[varKey]
		tlp.wp.Lib.Path = values[pathKey]
	case strings.Contains(text, "+akogen:ExternalType:"):
		values, err := parseAnnotationCSVValue(text)
		if err != nil {
			return fmt.Errorf("failed to parse ExternalType annotation: %w", err)
		}
		tlp.externalAlias = values[varKey]
		tlp.externalType = values[typeKey]
	case strings.Contains(text, "+akogen:ExternalAPI:"):
		values, err := parseAnnotationCSVValue(text)
		if err != nil {
			return fmt.Errorf("failed to parse ExternalAPI annotation: %w", err)
		}
		tlp.externalAPIAlias = values[varKey]
		tlp.externalAPIType = values[typeKey]
	case strings.Contains(text, "+akogen:WrapperType:"):
		values, err := parseAnnotationCSVValue(text)
		if err != nil {
			return fmt.Errorf("failed to parse WrapperType annotation: %w", err)
		}
		tlp.wp.Wrapper = NewNamedType(values[varKey], values[typeKey])
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

	switch node.(type) {
	case *ast.Comment:
		c := node.(*ast.Comment)
		if strings.Contains(c.Text, "+akogen:InternalType:") {
			values, err := parseAnnotationCSVValue(c.Text)
			if err != nil {
				tlp.err = fmt.Errorf("failed to parse InternalType annotation")
				return false
			}
			var ok bool
			tlp.internalAlias, ok = values[varKey]
			if !ok {
				tlp.err = fmt.Errorf("missing var for InternalType")
				return false
			}
			tlp.internalPointer = isTrue(values[pointerKey])
			return true
		}
		if err := tlp.parseComment(c); err != nil {
			tlp.err = err
			return false
		}
	case *ast.TypeSpec:
		ts := (node).(*ast.TypeSpec)
		if err := tlp.parseInternalType(ts); err != nil {
			tlp.err = err
			return false
		}
	default:
		if tlp.internalAlias != "" {
			tlp.internalAlias = ""
		}
		return true
	}

	if err := tlp.setExternal(); err != nil {
		tlp.err = err
		return false
	}
	if err := tlp.setExternalAPI(); err != nil {
		tlp.err = err
		return false
	}
	return true
}

func (tlp *translatorLayerParser) parseInternalType(ts *ast.TypeSpec) error {
	if tlp.internalAlias == "" {
		return nil
	}
	typeName := ts.Name.Name
	if tlp.internalPointer {
		typeName = "*" + typeName
	}
	tlp.wp.Internal = NewStruct(NewNamedType(tlp.internalAlias, typeName))
	return nil
}

func (tlp *translatorLayerParser) setExternal() error {
	if tlp.externalType == "" {
		return fmt.Errorf("missing type value for ExternalType annotation")
	}
	externalNamedType, err := parseNamedType(tlp.externalAlias, tlp.externalType, tlp.wp.Lib.Path)
	if err != nil {
		return fmt.Errorf("failed to parse ExternalType named type: %w", err)
	}
	tlp.wp.External = NewStruct(externalNamedType)
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
		return "", "", fmt.Errorf("failed to extract assignment value from %s", s)
	}
	return parts[0], strings.Trim(parts[1], " \""), nil
}

func parseAnnotationValue(s string) (string, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("failed to extract annotation value from %s", s)
	}
	return parts[2], nil
}

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

func parseNamedType(alias, namespacedTypeName, pkgPath string) (NamedType, error) {
	dereferenced, isPtr := dereference(namespacedTypeName)
	_, typeName, err := splitNamespacedType(dereferenced)
	if err != nil {
		return NamedType{}, fmt.Errorf("failed to parse ExternalType named type: %w", err)
	}
	fullTypeName := fmt.Sprintf("%s.%s", pkgPath, typeName)
	if isPtr {
		fullTypeName = fmt.Sprintf("*%s", fullTypeName)
	}
	return NewNamedType(alias, fullTypeName), nil
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

func isTrue(s string) bool {
	return strings.TrimSpace(strings.ToLower(s)) == "true"
}
