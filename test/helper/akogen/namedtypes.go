package akogen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

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

func (nt NamedType) StripPackageAndName(pkgPath string) NamedType {
	nt.Type = nt.Type.StripPackage(pkgPath)
	nt.Name = removeBase(nt.Name, pkgPath)
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

func (nt NamedType) generateZeroValue() *jen.Statement {
	if nt.Primitive == nil {
		return nt.Type.generateZeroValue()
	}
	return nt.Primitive.generateZeroValue()
}

func (nt NamedType) assignableFrom(other NamedType) bool {
	nonPtrType := nt.Type.dereference()
	return other.Type.dereference() == nonPtrType.dereference() ||
		(other.Primitive != nil && nonPtrType.dereference() == *other.Primitive) ||
		(nt.Primitive != nil && other.Type.dereference() == *nt.Primitive)
}

func (nt NamedType) generateMethodReceiver() *jen.Statement {
	return nt.generateNameType()
}

func (nt NamedType) generateNameType() *jen.Statement {
	return nt.Type.generate(jen.Id(nt.Name))
}

type NamedTypes []NamedType

func (nts NamedTypes) replaceType(original, replacement NamedType) NamedTypes {
	if len(nts) < 1 {
		panic("expected one or more named types")
	}
	list := make(NamedTypes, 0, len(nts))
	for _, nt := range nts {
		if nt == original {
			nt = replacement
		}
		list = append(list, nt)
	}
	return list
}

func (nts NamedTypes) generateArgsSignature() []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, nt.generateNameType())
	}
	return list
}

func (nts NamedTypes) generateReturnsSignature() []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, nt.Type.generate(nil))
	}
	return list
}

func (nts NamedTypes) generateCallArgs() []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, jen.Id(nt.Name))
	}
	return list
}

func (nts NamedTypes) generateList() *jen.Statement {
	return jen.List(nts.generateCallArgs()...)
}

func (nts NamedTypes) generateAssignCallReturns() *jen.Statement {
	return nts.generateList().Op(":=")
}

func (nts NamedTypes) generateReturnError() *jen.Statement {
	if len(nts) < 1 {
		panic("expected one or more returns")
	}
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		item := nt.generateZeroValue()
		if nt.Type == "error" {
			item = jen.Id(nt.Name)
		}
		list = append(list, item)
	}
	return jen.List(list...)
}
