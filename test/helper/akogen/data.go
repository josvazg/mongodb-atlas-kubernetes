package akogen

import "fmt"

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
