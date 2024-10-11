package metadata

import (
	"fmt"
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

func (nt NamedType) Dereference() NamedType {
	return NamedType{
		Type:      nt.Type.Dereference(),
		Name:      nt.Name,
		Primitive: nt.Primitive,
	}
}

func (nt NamedType) Pointer() NamedType {
	return NamedType{
		Type:      nt.Type.Pointer(),
		Name:      nt.Name,
		Primitive: nt.Primitive,
	}
}

func (nt NamedType) AsPrimitive() (NamedType, bool) {
	if nt.Primitive == nil {
		return nt, false
	}
	newType := *nt.Primitive
	if nt.IsPointer() {
		newType = newType.Pointer()
	}
	return NamedType{
		Type:      newType,
		Name:      nt.Name,
		Primitive: nil,
	}, true
}

func (nt NamedType) AssignableFrom(other NamedType) bool {
	nonPtrType := nt.Type.Dereference()
	return other.Type.Dereference() == nonPtrType.Dereference() ||
		(other.Primitive != nil && nonPtrType.Dereference() == *other.Primitive) ||
		(nt.Primitive != nil && other.Type.Dereference() == *nt.Primitive)
}

type NamedTypes []NamedType

func (nts NamedTypes) ReplaceType(original, replacement NamedType) NamedTypes {
	return nts.replaceType(original, replacement)
}

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
