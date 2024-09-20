package akogen

import (
	"strings"

	"github.com/dave/jennifer/jen"
)

type Type string

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

func (t Type) dereference() string {
	str := string(t)
	for str[0] == '*' {
		str = str[1:]
	}
	return str
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
	// TODO be smarter about non primitive types?
	return jen.Nil()
}

type NamedType struct {
	Type
	Name string
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
		item := &jen.Statement{}
		if nt.Type == "error" {
			item = jen.Id(nt.Name)
		} else {
			item = nt.zeroValue()
		}
		list = append(list, item)
	}
	return jen.List(list...)
}
