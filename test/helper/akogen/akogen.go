package akogen

import (
	"github.com/dave/jennifer/jen"
)

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

func generateFunctionSignature(f *jen.File, fns *FunctionSignature, blockStatements ...jen.Code) *jen.Statement {
	return f.Func().Id(fns.Name).Params(fns.Args.generateArgsSignature()...).
		Params(fns.Returns.generateReturnsSignature()...).Block(blockStatements...)
}

func generateMethodSignature(f *jen.File, m *MethodSignature, blockStatements ...jen.Code) *jen.Statement {
	return f.Func().Params(m.Receiver.generateMethodReceiver()).Id(m.Name).
		Params(m.Args.generateArgsSignature()...).
		Params(m.Returns.generateReturnsSignature()...).Block(blockStatements...)
}

func generateReturnOnError(returns NamedTypes) *jen.Statement {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	return jen.If(jen.Id("err").Op("!=").Nil()).Block(
		jen.Return(returns.generateReturnError()),
	)
}

func generateReturns(returns NamedTypes) *jen.Statement {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	list := make([]jen.Code, 0, len(returns))
	for _, ret := range returns {
		list = append(list, jen.Id(ret.Name))
	}
	return jen.Return(jen.List(list...))
}
