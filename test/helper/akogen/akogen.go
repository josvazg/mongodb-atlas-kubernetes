package akogen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

type TranslationLayer struct {
	PackageName string
	Type        Type
	Translation Translation
	WrappedType *WrappedType
}

type Translation struct {
	Internal     NamedType
	External     NamedType
	ExternalName string
}

type WrappedType struct {
	NamedType
	Lib            Import
	WrapperMethods []WrapperMethod
}

type WrapperMethod struct {
	MethodSignature
	WrappedCall CallSignature
}

type Type string

type Import struct {
	Alias, Path string
}

type NamedType struct {
	Name string
	Type Type
}

type CallSignature struct {
	Name    string
	Args    []NamedType
	Returns []NamedType
}

type MethodSignature struct {
	CallSignature
	ImplType NamedType
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
		f.Type().Id(dereference(tl.Type)).Struct(
			jen.Id(tl.WrappedType.Name).Qual(tl.WrappedType.Lib.Path, string(tl.WrappedType.Type)),
		)
		addedFunc := false
		for _, wm := range tl.WrappedType.WrapperMethods {
			if addedFunc {
				f.Empty()
			}
			addMethodSignature(
				f,
				&wm.MethodSignature,
				wrapAPICall(&wm, tl.WrappedType.Name, &tl.Translation),
				returnOnError(wm.Returns),
				returns(translateArgs(&tl.Translation, wm.WrappedCall.Returns)),
			)
			addedFunc = true
		}
	}
	return f.GoString(), nil
}

func isEmpty(tl *TranslationLayer) bool {
	return tl.PackageName == "" && tl.WrappedType == nil
}

func dereference(t Type) string {
	str := string(t)
	for str[0] == '*' {
		str = str[1:]
	}
	return str
}

func addMethodSignature(f *jen.File, m *MethodSignature, blockStatements ...jen.Code) *jen.Statement {
	return f.Func().Params(
		methodReceiver(m.ImplType)).Id(m.Name).
		Params(argsSignature(m.Args)...).
		Params(returnsSignature(m.Returns)...).Block(blockStatements...)
}

func wrapAPICall(wm *WrapperMethod, fieldName string, translation *Translation) *jen.Statement {
	return assignCallReturns(wm.WrappedCall.Returns).Id(wm.ImplType.Name).Dot(fieldName).Dot(wm.WrappedCall.Name).
		Call(callArgs(translateArgs(translation, wm.Args))...)
}

func translateArgs(translation *Translation, vars []NamedType) []NamedType {
	outVars := make([]NamedType, 0, len(vars))
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

func methodReceiver(nt NamedType) jen.Code {
	return nameType(nt)
}

func nameType(nt NamedType) jen.Code {
	return jen.Id(nt.Name).Id(string(nt.Type))
}

func argsSignature(args []NamedType) []jen.Code {
	list := []jen.Code{}
	for _, arg := range args {
		list = append(list, nameType(arg))
	}
	return list
}

func returnsSignature(returns []NamedType) []jen.Code {
	list := []jen.Code{}
	for _, ret := range returns {
		list = append(list, jen.Id(string(ret.Type)))
	}
	return list
}

func callArgs(args []NamedType) []jen.Code {
	list := []jen.Code{}
	for _, arg := range args {
		list = append(list, jen.Id(arg.Name))
	}
	return list
}

func list(vars []NamedType) *jen.Statement {
	list := make([]jen.Code, 0, len(vars))
	for _, ret := range vars {
		list = append(list, jen.Id(ret.Name))
	}
	return jen.List(list...)
}

func assignCallReturns(returns []NamedType) *jen.Statement {
	return list(returns).Op(":=")
}

func returnOnError(returns []NamedType) jen.Code {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	return jen.If(jen.Id("err").Op("!=").Nil()).Block(
		jen.Return(returnError(returns)),
	)
}

func returnError(returns []NamedType) *jen.Statement {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	list := make([]jen.Code, 0, len(returns))
	for _, ret := range returns {
		item := &jen.Statement{}
		if ret.Type == "error" {
			item = jen.Id(ret.Name)
		} else {
			item = zeroValueOf(ret.Type)
		}
		list = append(list, item)
	}
	return jen.List(list...)
}

func returns(returns []NamedType) *jen.Statement {
	if len(returns) < 1 {
		panic("expected one or more returns")
	}
	list := make([]jen.Code, 0, len(returns))
	for _, ret := range returns {
		list = append(list, jen.Id(ret.Name))
	}
	return jen.Return(jen.List(list...))
}

func zeroValueOf(t Type) *jen.Statement {
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
