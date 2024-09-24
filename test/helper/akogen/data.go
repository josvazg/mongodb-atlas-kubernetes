package akogen

type DataKind int

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

type DataField struct {
	DataType
	FieldName string
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
