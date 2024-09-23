package akogen

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"
)

const (
	pointerLib = "github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/pointer"
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
	if tl.isEmpty() {
		return "", ErrEmptySpec
	}
	f := jen.NewFile(tl.PackageName)
	if tl.WrappedType != nil {
		tl.defineWrapperType(f)
		for _, wm := range tl.WrappedType.WrapperMethods {
			tl.implementAPICallWrapping(f, wm)
			f.Empty()
		}
		if err := tl.defineToExternalConversion(f); err != nil {
			return "", fmt.Errorf("failed to generate conversion to external type: %w", err)
		}
		f.Empty()
		if err := tl.defineFromExternalConversion(f); err != nil {
			return "", fmt.Errorf("failed to generate conversion from external type: %w", err)
		}
	}
	return f.GoString(), nil
}

func (tl *TranslationLayer) defineWrapperType(f *jen.File) {
	f.ImportName(tl.WrappedType.Lib.Path, tl.WrappedType.Lib.Alias)
	f.Type().Id(tl.WrappedType.Wrapper.dereference().Type.String()).Struct(
		jen.Id(tl.WrappedType.ExternalAPI.Name).Qual(
			tl.WrappedType.Lib.Path, string(tl.WrappedType.ExternalAPI.Type),
		),
	)
}

func (tl *TranslationLayer) implementAPICallWrapping(f *jen.File, wm WrapperMethod) {
	addMethodSignature(
		f,
		&wm.MethodSignature,
		wrapAPICall(&wm, &tl.WrappedType.Translation),
		returnOnError(wm.Returns),
		returns(translateArgs(&tl.WrappedType.Translation, wm.WrappedCall.Returns)),
	)
}

func (tl *TranslationLayer) defineToExternalConversion(f *jen.File) error {
	conversion, err := returnConversion(&tl.WrappedType.External, &tl.WrappedType.Internal)
	if err != nil {
		return fmt.Errorf("struct conversion failed: %v", err)
	}
	addFunctionSignature(
		f,
		&FunctionSignature{
			Name:    fmt.Sprintf("to%s", tl.WrappedType.ExternalName),
			Args:    []NamedType{tl.WrappedType.Internal.NamedType},
			Returns: []NamedType{tl.WrappedType.External.NamedType},
		},
		conversion,
	)
	return nil
}

func (tl *TranslationLayer) defineFromExternalConversion(f *jen.File) error {
	conversion, err := returnConversion(&tl.WrappedType.Internal, &tl.WrappedType.External)
	if err != nil {
		return fmt.Errorf("struct conversion failed: %v", err)
	}
	addFunctionSignature(
		f,
		&FunctionSignature{
			Name:    fmt.Sprintf("from%s", tl.WrappedType.ExternalName),
			Args:    []NamedType{tl.WrappedType.External.NamedType},
			Returns: []NamedType{tl.WrappedType.Internal.NamedType},
		},
		conversion,
	)
	return nil
}

func (tl *TranslationLayer) isEmpty() bool {
	return tl.PackageName == "" && tl.WrappedType == nil
}

func addFunctionSignature(f *jen.File, fns *FunctionSignature, blockStatements ...jen.Code) *jen.Statement {
	return f.Func().Id(fns.Name).Params(fns.Args.argsSignature()...).
		Params(fns.Returns.returnsSignature()...).Block(blockStatements...)
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

func returnConversion(dst, src *Struct) (*jen.Statement, error) {
	typeConversion, err := convertTypes(dst, src)
	if err != nil {
		return nil, err
	}
	return jen.Return(typeConversion), nil
}

func convertTypes(dst, src *Struct) (*jen.Statement, error) {
	fieldConversions, err := convertFields(dst, src)
	if err != nil {
		return nil, err
	}
	return jen.Op("&").Id(string(dst.Type.dereference())).Values(fieldConversions), nil
}

func convertFields(dst, src *Struct) (jen.Dict, error) {
	remaining := src.Fields
	values := jen.Dict{}
	for _, field := range dst.Fields {
		var err error
		var conversion *jen.Statement
		remaining, conversion, err = computeConversion(remaining, src, field)
		if err != nil {
			return nil, fmt.Errorf("failed to compute conversion: %w", err)
		}
		values[jen.Id(field.Name)] = conversion
	}
	return values, nil
}

func computeConversion(candidates NamedTypes, src *Struct, target NamedType) (NamedTypes, *jen.Statement, error) {
	remaining, srcField, err := findSourceField(candidates, target)
	if err != nil {
		return remaining, nil, fmt.Errorf("could not compute conversion: %w", err)
	}
	if srcField == nil {
		panic(fmt.Sprintf("findResourceField must report error when source is not found for target %v at %v",
			target, candidates))
	}
	assignment, err := computeAssignment(target, src, *srcField)
	if err != nil {
		return remaining, nil, fmt.Errorf("could not compute conversion assignment: %w", err)
	}
	return remaining, assignment, nil
}

func findSourceField(candidates NamedTypes, target NamedType) (NamedTypes, *NamedType, error) {
	prefix := NamedTypes{}
	for i, candidate := range candidates {
		if strings.EqualFold(target.Name, candidate.Name) && target.assignableFrom(candidate) {
			remaining := append(prefix, candidates[i+1:]...)
			if reflect.DeepEqual(remaining, candidates) {
				panic(fmt.Sprintf("remaining cannot match candidates, source %v has not been extracted from %v",
					candidate, remaining))
			}
			return remaining, &candidate, nil
		}
		prefix = append(prefix, candidate)
	}
	return nil, nil, fmt.Errorf("could not find corresponding field for %v at %v", target, candidates)
}

func computeAssignment(target NamedType, src *Struct, srcField NamedType) (*jen.Statement, error) {
	switch {
	// Same types on both sides
	case srcField.Type == target.Type:
		return jen.Id(src.Name).Dot(srcField.Name), nil
	// source field can be converted to the primitive type of the target
	case srcField.Primitive != nil && *srcField.Primitive == target.Type.dereference():
		primitive, ok := srcField.primitive()
		if !ok {
			return nil, fmt.Errorf("could not cast field %v.%v to primitive type", src, srcField)
		}
		assignment, err := computeAssignment(target, src, primitive)
		if err != nil {
			return nil, fmt.Errorf("failed to assign after casting field to primitive: %w", err)
		}
		return jen.Id(srcField.Type.dereference().String()).Call(assignment), nil
	// Target is a pointer
	case target.isPointer() && !srcField.isPointer():
		assignment, err := computeAssignment(target, src, srcField.pointer())
		if err != nil {
			return nil, fmt.Errorf("failed to assign to pointer target %v: %w", target, err)
		}
		return jen.Qual(pointerLib, "MakePtr").Call(assignment), nil
	// Source field is the pointer
	case srcField.isPointer() && !target.isPointer():
		deref := srcField.dereference()
		assignment, err := computeAssignment(target, src, deref)
		if err != nil {
			return nil, fmt.Errorf("failed to assign to pointer target %v: %w", target, err)
		}
		return jen.Qual(pointerLib, "GetOrDefault").Call(jen.List(assignment, deref.zeroValue())), nil
	// target can be converted from the primitive type of the source field
	case target.Primitive != nil && *target.Primitive == srcField.Type.dereference():
		primitive, ok := target.primitive()
		if !ok {
			return nil, fmt.Errorf("could not cast target %v to primitive type", target)
		}
		assignment, err := computeAssignment(primitive, src, srcField)
		if err != nil {
			return nil, fmt.Errorf("failed to assign after casting field to primitive: %w", err)
		}
		return jen.Id(primitive.Type.dereference().String()).Call(assignment), nil
	default:
		return nil, fmt.Errorf("cannot find way to assign %s.%v to %v", src, srcField, target)
	}
}
