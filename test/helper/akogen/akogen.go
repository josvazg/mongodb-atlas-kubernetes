package akogen

import (
	"errors"
	"strings"

	"github.com/dave/jennifer/jen"
)

type TranslationLayer struct {
	PackageName  string
	CallWrappers []WrappedCall
}

type WrappedCall struct {
	MethodSignature
	WrappedLib  Import
	WrappedType NamedType
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
	defined := map[Type]struct{}{}
	for _, wm := range tl.CallWrappers {
		implType := wm.MethodSignature.ImplType.Type
		if !isDefined(defined, implType) {
			f.ImportName(wm.WrappedLib.Path, wm.WrappedLib.Alias)
			f.Type().Id(dereference(implType)).Struct(
				jen.Id(wm.WrappedType.Name).Qual(wm.WrappedLib.Path, string(wm.WrappedType.Type)),
			)
			define(defined, implType)
		}
		addMethodSignature(
			f,
			wm.MethodSignature,
			wrapAPICall(wm.ImplType, wm.WrappedType.Name, wm.WrappedCall),
			returnOnError(wm.Returns),
			jen.Return(
				jen.Id("fromAtlas").Params(jen.Id(wm.Returns[0].Name)),
				jen.Nil(),
			),
		)
	}
	return f.GoString(), nil
}

func isEmpty(tl *TranslationLayer) bool {
	return tl.PackageName == "" && len(tl.CallWrappers) == 0
}

func isDefined(defined map[Type]struct{}, t Type) bool {
	_, ok := defined[t]
	return ok
}

func define(defined map[Type]struct{}, t Type) {
	defined[t] = struct{}{}
}

func dereference(t Type) string {
	str := string(t)
	for str[0] == '*' {
		str = str[1:]
	}
	return str
}

func addMethodSignature(f *jen.File, m MethodSignature, blockStatements ...jen.Code) *jen.Statement {
	return f.Func().Params(
		methodReceiver(m.ImplType)).Id(m.Name).
		Params(argsSignature(m.Args)...).
		Params(returnsSignature(m.Returns)...).Block(blockStatements...)
}

func wrapAPICall(this NamedType, fieldName string, wrappedCall CallSignature) *jen.Statement {
	return assignCallReturns(wrappedCall.Returns).Id(this.Name).Dot(fieldName).Dot(wrappedCall.Name).
		Call(callArgs(wrappedCall.Args)...)
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
