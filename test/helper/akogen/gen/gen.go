package gen

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/metadata"
)

func Type(t metadata.Type) *jen.Statement {
	lib := t.Lib()
	if lib != "" {
		if t.IsPointer() {
			return jen.Op("*").Qual(lib, t.Base())
		}
		return jen.Qual(lib, t.Base())
	}
	return jen.Id(string(t))
}

func AddType(prior *jen.Statement, t metadata.Type) *jen.Statement {
	if prior == nil {
		prior = jen.Empty()
	}
	lib := t.Lib()
	if lib != "" {
		if t.IsPointer() {
			return prior.Op("*").Qual(lib, t.Base())
		}
		return prior.Qual(lib, t.Base())
	}
	return prior.Id(string(t))
}

func Function(f *jen.File, fns *metadata.FunctionSignature) *jen.Statement {
	return f.Func().Id(fns.Name).Params(argsSignature(fns.Args)...).Params(returnsSignature(fns.Returns)...)
}

func Method(f *jen.File, m *metadata.MethodSignature) *jen.Statement {
	return f.Func().Params(methodReceiver(m.Receiver)).Id(m.Name).Params(argsSignature(m.Args)...).
		Params(returnsSignature(m.Returns)...)
}

func Returns(returns metadata.NamedTypes) *jen.Statement {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	list := make([]jen.Code, 0, len(returns))
	for _, ret := range returns {
		list = append(list, jen.Id(ret.Name))
	}
	return jen.Return(jen.List(list...))
}

func ReturnOnError(returns metadata.NamedTypes) *jen.Statement {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	return jen.If(jen.Id("err").Op("!=").Nil()).Block(
		jen.Return(returnError(returns)),
	)
}

func AssignCallReturns(nts metadata.NamedTypes) *jen.Statement {
	return list(nts).Op(":=")
}

func CallArgs(nts metadata.NamedTypes) []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, jen.Id(nt.Name))
	}
	return list
}

func ZeroValue(nt metadata.NamedType) *jen.Statement {
	if nt.Primitive == nil {
		return zeroValueType(nt.Type)
	}
	return zeroValueType(*nt.Primitive)
}

func methodReceiver(nt metadata.NamedType) *jen.Statement {
	return nameType(nt)
}

func nameType(nt metadata.NamedType) *jen.Statement {
	return AddType(jen.Id(nt.Name), nt.Type)
}

func argsSignature(nts metadata.NamedTypes) []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, nameType(nt))
	}
	return list
}

func returnsSignature(nts metadata.NamedTypes) []jen.Code {
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		list = append(list, Type(nt.Type))
	}
	return list
}

func list(nts metadata.NamedTypes) *jen.Statement {
	return jen.List(CallArgs(nts)...)
}

func returnError(nts metadata.NamedTypes) *jen.Statement {
	if len(nts) < 1 {
		panic("expected one or more returns")
	}
	list := make([]jen.Code, 0, len(nts))
	for _, nt := range nts {
		item := ZeroValue(nt)
		if nt.Type == "error" {
			item = jen.Id(nt.Name)
		}
		list = append(list, item)
	}
	return jen.List(list...)
}

func zeroValueType(t metadata.Type) *jen.Statement {
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
