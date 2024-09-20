package akogen

import (
	"errors"
	"fmt"

	"github.com/dave/jennifer/jen"
)

type TranslationLayer struct {
	PackageName string
	WrappedType *WrappedType
}

type WrappedType struct {
	Translation
	WrapperMethods []WrapperMethod
}

type Translation struct {
	Lib          Import
	ExternalName string
	External     Struct
	ExternalAPI  NamedType
	Internal     Struct
	Wrapper      NamedType
}

type WrapperMethod struct {
	MethodSignature
	WrappedCall FunctionSignature
}

var (
	// ErrorNilSpec code spec cannot be nil
	ErrNilSpec = errors.New("code specification cannot be nil")

	// ErrorEmptySpec code spec cannot be empty
	ErrEmptySpec = errors.New("code specification cannot be empty")
)

func GenerateTranslationLayer(tl *TranslationLayer) (string, error) {
	if tl == nil {
		return "", ErrNilSpec
	}
	if isEmpty(tl) {
		return "", ErrEmptySpec
	}
	f := jen.NewFile(tl.PackageName)
	if tl.WrappedType != nil {
		f.ImportName(tl.WrappedType.Lib.Path, tl.WrappedType.Lib.Alias)
		f.Type().Id(tl.WrappedType.Wrapper.dereference()).Struct(
			jen.Id(tl.WrappedType.ExternalAPI.Name).Qual(
				tl.WrappedType.Lib.Path, string(tl.WrappedType.ExternalAPI.Type),
			),
		)
		addedFunc := false
		for _, wm := range tl.WrappedType.WrapperMethods {
			if addedFunc {
				f.Empty()
			}
			addMethodSignature(
				f,
				&wm.MethodSignature,
				wrapAPICall(&wm, &tl.WrappedType.Translation),
				returnOnError(wm.Returns),
				returns(translateArgs(&tl.WrappedType.Translation, wm.WrappedCall.Returns)),
			)
			addedFunc = true
		}
	}
	return f.GoString(), nil
}

func isEmpty(tl *TranslationLayer) bool {
	return tl.PackageName == "" && tl.WrappedType == nil
}

func addMethodSignature(f *jen.File, m *MethodSignature, blockStatements ...jen.Code) *jen.Statement {
	return f.Func().Params(m.Receiver.methodReceiver()).Id(m.Name).
		Params(m.Args.argsSignature()...).
		Params(m.Returns.returnsSignature()...).Block(blockStatements...)
}

func wrapAPICall(wm *WrapperMethod, translation *Translation) *jen.Statement {
	return wm.WrappedCall.Returns.assignCallReturns().
		Id(wm.Receiver.Name).Dot(translation.ExternalAPI.Name).Dot(wm.WrappedCall.Name).
		Call(translateArgs(translation, wm.Args).callArgs()...)
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

func returnOnError(returns NamedTypes) jen.Code {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	return jen.If(jen.Id("err").Op("!=").Nil()).Block(
		jen.Return(returns.returnError()),
	)
}

func returns(returns NamedTypes) *jen.Statement {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	list := make([]jen.Code, 0, len(returns))
	for _, ret := range returns {
		list = append(list, jen.Id(ret.Name))
	}
	return jen.Return(jen.List(list...))
}
