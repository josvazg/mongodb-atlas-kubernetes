package akogen

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/metadata"
)

const (
	// Auto is a setting that gets automatically overridden by a sane default
	Auto = ""

	// DefaultExternalName is the default external name
	DefaultExternalName = "ExternalSystem"

	// DefaultExternalField name
	DefaultExternalField = "external"

	// DefaultInternalField name
	DefaultInternalField = "internal"
)

var DefaultSettings = TranslationLayerSettings{
	ExternalName: DefaultExternalName,
	ImportAlias:  Auto,
	WrapperType:  Auto,
}

type TranslationLayerSpec struct {
	PackageName  string
	Name         string
	API          reflect.Type
	ExternalType any
	InternalType any
}

type TranslationLayerSettings struct {
	ExternalName string
	ImportAlias  string
	WrapperType  string
}

func NewTranslationLayer(tls *TranslationLayerSpec, settings TranslationLayerSettings) *TranslationLayer {
	if tls.API.Kind() != reflect.Interface {
		panic(fmt.Sprintf("API expected to be an Interface, but got %v", tls.API.Kind()))
	}
	pkgPath := tls.API.PkgPath()
	internalType := reflect.TypeOf(tls.InternalType)
	localPkgPath := internalType.PkgPath()
	if internalType.Kind() == reflect.Pointer {
		localPkgPath = internalType.Elem().PkgPath()
	}
	importAlias := importAlias(settings.ImportAlias, pkgPath)
	wrapperTypeName := wrapper(settings.WrapperType, tls.Name)
	external := NewDataTypeFromReflect(reflect.TypeOf(tls.ExternalType))
	internal := NewDataTypeFromReflect(reflect.TypeOf(tls.InternalType)).StripLocalPackage(localPkgPath)
	return &TranslationLayer{
		PackageName: tls.PackageName,
		WrappedType: &WrappedType{
			Translation: Translation{
				Lib:          metadata.Import{Alias: importAlias, Path: pkgPath},
				ExternalName: settings.ExternalName,
				External:     external,
				ExternalAPI:  metadata.NewNamedType(shortenName(tls.API.Name()), tls.API.Name()),
				Internal:     internal,
				Wrapper:      metadata.NewNamedType(shortenName(wrapperTypeName), wrapperTypeName),
			},
			WrapperMethods: NewWrapperMethodsFromReflect(
				settings.ExternalName,
				metadata.NewNamedType(shortenName(settings.WrapperType), settings.WrapperType).Pointer(),
				tls.API,
				external.NamedType,
				internal.NamedType,
			),
		},
	}
}

func importAlias(alias, pkgPath string) string {
	if alias == Auto {
		base := filepath.Base(pkgPath)
		return base
	}
	return alias
}

func wrapper(wrapperName, fallback string) string {
	if wrapperName == Auto {
		return fallback
	}
	return wrapperName
}

func NewTypeFromReflect(t reflect.Type) metadata.Type {
	if t.PkgPath() != "" {
		return metadata.Type(fmt.Sprintf("%s.%s", t.PkgPath(), t.Name()))
	}
	return metadata.Type(t.Name())
}

func NewNamedTypeFromReflect(name string, t reflect.Type) metadata.NamedType {
	var primitive *metadata.Type
	typeName := NewTypeFromReflect(t)
	if isPrimitive(t) {
		p := metadata.Type(t.Kind().String())
		if p != typeName {
			primitive = &p
		}
	}
	return metadata.NamedType{
		Name:      name,
		Type:      typeName,
		Primitive: primitive,
	}
}

func NewDataTypeFromReflect(t reflect.Type) *metadata.DataType {
	switch t.Kind() {
	case reflect.Struct:
		return newStructFromReflect(shortenTypeName(t), t)
	case reflect.Pointer:
		pointedType := t.Elem()
		dt := NewDataTypeFromReflect(pointedType)
		dt.NamedType.Type = dt.NamedType.Type.Pointer()
		return dt
	}
	return nil
}

func NewWrapperMethodsFromReflect(externalName string, wrapper metadata.NamedType, apiType reflect.Type, external, internal metadata.NamedType) []WrapperMethod {
	wms := make([]WrapperMethod, 0, apiType.NumMethod())
	for i := 0; i < apiType.NumMethod(); i++ {
		m := apiType.Method(i)
		wms = append(wms, WrapperMethod{
			MethodSignature: metadata.MethodSignature{
				Receiver: wrapper,
				FunctionSignature: metadata.FunctionSignature{
					Name:    m.Name,
					Args:    argsFromReflect(m.Type).ReplaceType(external, internal),
					Returns: returnsFromReflect(m.Type).ReplaceType(external, internal),
				},
			},
			WrappedCall: metadata.FunctionSignature{
				Name:    m.Name,
				Args:    argsFromReflect(m.Type),
				Returns: returnsFromReflect(m.Type),
			},
		})
	}
	return wms
}

var ArgsFromReflect = argsFromReflect

func argsFromReflect(t reflect.Type) metadata.NamedTypes {
	args := make([]metadata.NamedType, 0, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		nt := NewNamedTypeFromReflect(shortenTypeName(argType), argType)
		if argType.Kind() == reflect.Pointer {
			argType = argType.Elem()
			nt = NewNamedTypeFromReflect(shortenTypeName(argType), argType).Pointer()
		}
		args = append(args, nt)
	}
	return args
}

var ReturnsFromReflect = returnsFromReflect

func returnsFromReflect(t reflect.Type) metadata.NamedTypes {
	returns := make([]metadata.NamedType, 0, t.NumOut())
	for i := 0; i < t.NumOut(); i++ {
		returnType := t.Out(i)
		nt := NewNamedTypeFromReflect(shortenTypeName(returnType), returnType)
		if returnType.Kind() == reflect.Pointer {
			returnType = returnType.Elem()
			nt = NewNamedTypeFromReflect(shortenTypeName(returnType), returnType).Pointer()
		}
		returns = append(returns, nt)
	}
	return returns
}

func newStructFromReflect(name string, st reflect.Type) *metadata.DataType {
	return &metadata.DataType{
		NamedType: NewNamedTypeFromReflect(name, st),
		Kind:      metadata.Struct,
		Fields:    dataFieldsFromReflect(st),
	}
}

func dataFieldsFromReflect(st reflect.Type) []*metadata.DataField {
	dataFields := []*metadata.DataField{}
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		if sf.IsExported() {
			dataField := newDataFieldFromReflect(sf.Name, sf.Type)
			dataFields = append(dataFields, dataField)
		}
	}
	return dataFields
}

func newDataFieldFromReflect(name string, t reflect.Type) *metadata.DataField {
	kind := t.Kind()
	switch {
	case kind == reflect.Struct:
		return newStructFieldFromReflect(name, t)
	case isPrimitive(t):
		return newSimpleFieldFromReflect(name, t)
	case kind == reflect.Pointer:
		pointedType := t.Elem()
		df := newDataFieldFromReflect(name, pointedType)
		df.NamedType.Type = df.NamedType.Type.Pointer()
		return df
	default:
		panic(fmt.Sprintf("unimplemented for Kind=%v", t.Kind()))
	}
}

func newStructFieldFromReflect(name string, t reflect.Type) *metadata.DataField {
	return &metadata.DataField{
		DataType:  *newStructFromReflect(shortenTypeName(t), t),
		FieldName: name,
	}
}

func newSimpleFieldFromReflect(name string, t reflect.Type) *metadata.DataField {
	return &metadata.DataField{
		DataType: metadata.DataType{
			NamedType: NewNamedTypeFromReflect(name, t),
			Kind:      metadata.SimpleField,
		},
		FieldName: name,
	}
}

func isPrimitive(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String:
		return true
	case reflect.Bool:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}

var ShortenTypeName = shortenTypeName

func shortenTypeName(t reflect.Type) string {
	base := ""
	if t != nil {
		base = filepath.Base(t.PkgPath())
	}
	return shorten(base, t.Name())
}
