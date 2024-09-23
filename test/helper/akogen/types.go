package akogen

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Type string

func (t Type) String() string {
	return string(t)
}

func (t Type) isPointer() bool {
	return t.String()[0] == '*'
}

type Import struct {
	Alias, Path string
}

type FunctionSignature struct {
	Name    string
	Args    NamedTypes
	Returns NamedTypes
}

type MethodSignature struct {
	FunctionSignature
	Receiver NamedType
}

type Struct struct {
	NamedType
	Fields NamedTypes
}

func (t Type) dereference() Type {
	for t.isPointer() {
		return Type(t.String()[1:])
	}
	return t
}

func (t Type) pointer() Type {
	return Type(fmt.Sprintf("*%v", t))
}

func (t Type) zeroValue() *jen.Statement {
	if t == "string" {
		return jen.Lit("")
	}
	if strings.HasPrefix(string(t), "*") {
		return jen.Nil()
	}
	if strings.Contains("int int8 int16 int32 int64 byte rune", string(t)) {
		return jen.Lit(0)
	}
	if strings.Contains("float float32 float64", string(t)) {
		return jen.Lit(0.0)
	}
	if strings.Contains("bool", string(t)) {
		return jen.False()
	}
	// TODO be smarter about non primitive types?
	return jen.Nil()
}

type NamedType struct {
	Type
	Name      string
	Primitive *Type
}

func NewNamedType(name, typeName string) NamedType {
	return NamedType{Name: name, Type: Type(typeName)}
}

func (nt NamedType) WithPrimitive(primitive Type) NamedType {
	nt.Primitive = &primitive
	return nt
}

func (nt NamedType) String() string {
	if nt.Primitive != nil {
		return fmt.Sprintf("{%s %v(%v)}", nt.Name, nt.Type, *nt.Primitive)
	}
	return fmt.Sprintf("{%s %v}", nt.Name, nt.Type)
}

func (nt NamedType) dereference() NamedType {
	return NamedType{
		Type:      nt.Type.dereference(),
		Name:      nt.Name,
		Primitive: nt.Primitive,
	}
}

func (nt NamedType) pointer() NamedType {
	return NamedType{
		Type:      nt.Type.pointer(),
		Name:      nt.Name,
		Primitive: nt.Primitive,
	}
}

func (nt NamedType) primitive() (NamedType, bool) {
	if nt.Primitive == nil {
		return nt, false
	}
	newType := *nt.Primitive
	if nt.isPointer() {
		newType = newType.pointer()
	}
	return NamedType{
		Type:      newType,
		Name:      nt.Name,
		Primitive: nil,
	}, true
}

func (nt NamedType) zeroValue() *jen.Statement {
	if nt.Primitive == nil {
		return nt.Type.zeroValue()
	}
	return nt.Primitive.zeroValue()
}

func (nt NamedType) assignableFrom(other NamedType) bool {
	nonPtrType := nt.Type.dereference()
	return nonPtrType.dereference() == nonPtrType.dereference() ||
	  (other.Primitive!= nil && nonPtrType.dereference() == *other.Primitive)
}

func (nt NamedType) methodReceiver() jen.Code {
	return nt.nameType()
}

func (nt NamedType) nameType() jen.Code {
	return jen.Id(nt.Name).Id(string(nt.Type))
}

type NamedTypes []NamedType

func (nts NamedTypes) argsSignature() []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, nt.nameType())
	}
	return list
}

func (nts NamedTypes) returnsSignature() []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, jen.Id(string(nt.Type)))
	}
	return list
}

func (nts NamedTypes) callArgs() []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, jen.Id(nt.Name))
	}
	return list
}

func (nts NamedTypes) list() *jen.Statement {
	return jen.List(nts.callArgs()...)
}

func (nts NamedTypes) assignCallReturns() *jen.Statement {
	return nts.list().Op(":=")
}

func (nts NamedTypes) returnError() *jen.Statement {
	if len(nts) < 1 {
		panic("expected one or more returns")
	}
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		item := nt.zeroValue()
		if nt.Type == "error" {
			item = jen.Id(nt.Name)
		}
		list = append(list, item)
	}
	return jen.List(list...)
}
