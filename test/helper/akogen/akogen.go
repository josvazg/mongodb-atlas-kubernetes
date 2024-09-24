package akogen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

func generateFunctionSignature(f *jen.File, fns *FunctionSignature, blockStatements ...jen.Code) *jen.Statement {
	return f.Func().Id(fns.Name).Params(fns.Args.argsSignature()...).
		Params(fns.Returns.returnsSignature()...).Block(blockStatements...)
}

func generateMethodSignature(f *jen.File, m *MethodSignature, blockStatements ...jen.Code) *jen.Statement {
	return f.Func().Params(m.Receiver.methodReceiver()).Id(m.Name).
		Params(m.Args.argsSignature()...).
		Params(m.Returns.returnsSignature()...).Block(blockStatements...)
}

func translateArgs(translation *Translation, vars []NamedType) NamedTypes {
	outVars := make(NamedTypes, 0, len(vars))
	for _, v := range vars {
		switch v.Type {
		case translation.External.Type:
			v.Name = fmt.Sprintf("from%s(%s)", translation.ExternalName, v.Name)
		case translation.Internal.Type:
			v.Name = fmt.Sprintf("to%s(%s)", translation.ExternalName, v.Name)
		case "error":
			v.Name = "nil"
		}
		outVars = append(outVars, v)
	}
	return outVars
}

func generateReturnOnError(returns NamedTypes) *jen.Statement {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	return jen.If(jen.Id("err").Op("!=").Nil()).Block(
		jen.Return(returns.returnError()),
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
