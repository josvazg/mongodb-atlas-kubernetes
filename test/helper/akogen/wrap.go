package akogen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type WrappedType struct {
	Translation
	WrapperMethods []WrapperMethod
}

type Translation struct {
	Lib          Import
	ExternalName string
	External     *DataType
	ExternalAPI  NamedType
	Internal     *DataType
	Wrapper      NamedType
}

type WrapperMethod struct {
	MethodSignature
	WrappedCall FunctionSignature
}

func (wt *WrappedType) generate(f *jen.File) error {
	if wt.External == nil || wt.Internal == nil {
		return fmt.Errorf("both wrapped internal and external types muts be set")
	}
	f.ImportName(wt.Lib.Path, wt.Lib.Alias)
	f.Type().Id(wt.Wrapper.dereference().Type.String()).Struct(
		jen.Id(wt.ExternalAPI.Name).Qual(
			wt.Lib.Path, string(wt.ExternalAPI.Type),
		),
	)
	for _, wm := range wt.WrapperMethods {
		wt.generateCallWrap(f, wm)
		f.Line()
	}
	return nil
}

func (wt *WrappedType) generateCallWrap(f *jen.File, wm WrapperMethod) {
	generateMethodSignature(
		f,
		&wm.MethodSignature,
		wrapCall(&wm, &wt.Translation),
		generateReturnOnError(wm.Returns),
		generateReturns(translateArgs(&wt.Translation, wm.WrappedCall.Returns)),
	)
}

func wrapCall(wm *WrapperMethod, translation *Translation) *jen.Statement {
	return wm.WrappedCall.Returns.generateAssignCallReturns().
		Id(wm.Receiver.Name).Dot(translation.ExternalAPI.Name).Dot(wm.WrappedCall.Name).
		Call(translateArgs(translation, wm.Args).generateCallArgs()...)
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
