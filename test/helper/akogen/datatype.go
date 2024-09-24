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
	Alias  string
	Fields []*DataType
}

func NewSimpleField(name, typeName string) *DataType {
	return &DataType{
		NamedType: NewNamedType(name, typeName),
		Kind:      SimpleField,
	}
}

func (ct *DataType) WithPrimitive(primitive Type) *DataType {
	ct.NamedType = ct.NamedType.WithPrimitive(primitive)
	return ct
}

func NewStruct(nt NamedType, fields ...*DataType) *DataType {
	return &DataType{
		NamedType: nt,
		Kind:      Struct,
		Alias:     nt.Name,
		Fields:    fields,
	}
}

func (ct *DataType) WithAlias(alias string) *DataType {
	ct.Alias = alias
	return ct
}

func (ct *DataType) primitive() (*DataType, bool) {
	if ct.Kind != SimpleField {
		return ct, false
	}
	nt, ok := ct.NamedType.primitive()
	if !ok {
		return ct, false
	}
	return &DataType{NamedType: nt, Kind: SimpleField}, true
}
