package akogen

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
)

type DataKind int

func (dk DataKind) String() string {
	switch dk {
	case SimpleField:
		return "SimpleField"
	case Struct:
		return "Struct"
	default:
		return "(unknown)"
	}
}

const (
	SimpleField DataKind = iota
	Struct
	// Array
	// Map
)

type DataType struct {
	NamedType
	Kind   DataKind
	Fields []*DataField
}

func (dt *DataType) String() string {
	return fmt.Sprintf("[%s Kind=%v %d Fields]", dt.NamedType, dt.Kind, len(dt.Fields))
}

func (df *DataType) StripLocalPackage(pkgPath string) *DataType {
	df.NamedType.Type = df.NamedType.Type.StripPackage(pkgPath)
	for _, field := range df.Fields {
		field.Name = removeBase(field.Name, pkgPath)
		field.StripLocalPackage(pkgPath)
	}
	return df
}

type DataField struct {
	DataType
	FieldName string
}

func (df *DataField) String() string {
	return fmt.Sprintf("%s:%s", df.FieldName, &df.DataType)
}

func (df *DataField) assignableFrom(other *DataField) bool {
	return df.Kind == other.Kind &&
		(df.Kind == SimpleField && df.NamedType.assignableFrom(other.NamedType)) ||
		(df.Kind != SimpleField)
}

func NewSimpleField(name, typeName string) *DataField {
	return &DataField{
		FieldName: name,
		DataType: DataType{
			NamedType: NewNamedType(name, typeName),
			Kind:      SimpleField,
		},
	}
}

func (df *DataField) WithPrimitive(primitive Type) *DataField {
	df.DataType.WithPrimitive(primitive)
	return df
}

func (dt *DataType) WithPrimitive(primitive Type) *DataType {
	dt.NamedType = dt.NamedType.WithPrimitive(primitive)
	return dt
}

func NewStruct(nt NamedType, fields ...*DataField) *DataType {
	return &DataType{
		NamedType: nt,
		Kind:      Struct,
		Fields:    fields,
	}
}

func NewStructField(fieldName string, nt NamedType, fields ...*DataField) *DataField {
	return &DataField{
		FieldName: fieldName,
		DataType:  *NewStruct(nt, fields...),
	}
}

func (dt *DataType) primitive() (*DataType, bool) {
	if dt.Kind != SimpleField {
		return dt, false
	}
	nt, ok := dt.NamedType.primitive()
	if !ok {
		return dt, false
	}
	return &DataType{NamedType: nt, Kind: SimpleField}, true
}

func NewDataTypeFromReflect(name string, t reflect.Type) *DataType {
	switch t.Kind() {
	case reflect.Struct:
		return newStructFromReflect(name, t)
	case reflect.Pointer:
		pointedType := t.Elem()
		dt := NewDataTypeFromReflect(name, pointedType)
		dt.NamedType.Type = dt.NamedType.Type.pointer()
		return dt
	}
	return nil
}

func newStructFromReflect(name string, st reflect.Type) *DataType {
	return &DataType{
		NamedType: NewNamedTypeFromReflect(name, st),
		Kind:      Struct,
		Fields:    dataFieldsFromReflect(st),
	}
}

func dataFieldsFromReflect(st reflect.Type) []*DataField {
	dataFields := []*DataField{}
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		if sf.IsExported() {
			dataField := newDataFieldFromReflect(sf.Name, sf.Type)
			dataFields = append(dataFields, dataField)
		}
	}
	return dataFields
}

func newDataFieldFromReflect(name string, t reflect.Type) *DataField {
	kind := t.Kind()
	switch {
	case kind == reflect.Struct:
		return newStructFieldFromReflect(name, t)
	case isPrimitive(t):
		return newSimpleFieldFromReflect(name, t)
	case kind == reflect.Pointer:
		pointedType := t.Elem()
		df := newDataFieldFromReflect(name, pointedType)
		df.NamedType.Type = df.NamedType.Type.pointer()
		return df
	default:
		panic(fmt.Sprintf("unimplemented for Kind=%v", t.Kind()))
	}
}

func newStructFieldFromReflect(name string, t reflect.Type) *DataField {
	return &DataField{
		DataType:  *newStructFromReflect(shorten(name, t), t),
		FieldName: name,
	}
}

var NewSimpleFieldFromReflect = newSimpleFieldFromReflect

func newSimpleFieldFromReflect(name string, t reflect.Type) *DataField {
	return &DataField{
		DataType:  DataType{
			NamedType: NewNamedTypeFromReflect(name, t),
			Kind:      SimpleField,
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

func shorten(s string, t reflect.Type) string {
	if len(s) == 0 {
		return ""
	}
	base := filepath.Base(t.PkgPath())
	filtered := string(s[0])
	for _, char := range s[1:] {
		if unicode.IsUpper(char) {
			filtered += string(unicode.ToLower(char))
		}
	}
	return fmt.Sprintf("%s%s", base, firstToUpper(filtered))
}

func removeBase(s, pkgPath string) string {
	if len(s) == 0 {
		return ""
	}
	base := filepath.Base(pkgPath)
	if strings.HasPrefix(s, base) && len(s) > len(base) {
		return firstToLower(strings.Replace(s, base, "", 1))
	}
	return s
}
