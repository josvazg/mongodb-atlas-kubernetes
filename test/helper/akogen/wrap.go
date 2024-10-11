package akogen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/gen"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/metadata"
)

type WrappedType struct {
	Translation
	WrapperMethods []WrapperMethod
}

type Translation struct {
	Lib          metadata.Import
	ExternalName string
	External     *metadata.DataType
	ExternalAPI  metadata.NamedType
	Internal     *metadata.DataType
	Wrapper      metadata.NamedType
}

type WrapperMethod struct {
	metadata.MethodSignature
	WrappedCall metadata.FunctionSignature
}

func (wt *WrappedType) generate(f *jen.File) error {
	if wt.External == nil || wt.Internal == nil {
		return fmt.Errorf("both wrapped internal and external types muts be set")
	}
	f.ImportName(wt.Lib.Path, wt.Lib.Alias)
	f.Type().Id(wt.Wrapper.Dereference().Type.String()).Struct(
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
	gen.Method(f, &wm.MethodSignature).Block(
		wrapCall(&wm, &wt.Translation),
		gen.ReturnOnError(wm.Returns),
		gen.Returns(translateArgs(&wt.Translation, wm.WrappedCall.Returns)),
	)
}

func wrapCall(wm *WrapperMethod, translation *Translation) *jen.Statement {
	return gen.AssignCallReturns(wm.WrappedCall.Returns).
		Id(wm.Receiver.Name).Dot(translation.ExternalAPI.Name).Dot(wm.WrappedCall.Name).
		Call(gen.CallArgs(translateArgs(translation, wm.Args))...)
}

func translateArgs(translation *Translation, vars []metadata.NamedType) metadata.NamedTypes {
	outVars := make(metadata.NamedTypes, 0, len(vars))
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
