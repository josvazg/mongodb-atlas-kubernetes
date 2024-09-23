package akogen

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

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
	External     *ComplexType
	ExternalAPI  NamedType
	Internal     *ComplexType
	Wrapper      NamedType
}

type WrapperMethod struct {
	MethodSignature
	WrappedCall FunctionSignature
}

type Conversion struct {
	Root   bool
	To     bool
	Name   string
	Source *ComplexType
	Target *ComplexType
}

func (c *Conversion) direction() string {
	if c.To {
		return "to"
	}
	return "from"
}

func (c *Conversion) method() string {
	if c.Root {
		return fmt.Sprintf("%s%s", c.direction(), c.Name)
	}
	return fmt.Sprintf("%s%s%s",
		firstToLower(c.Target.dereference().base()), firstToUpper(c.direction()), c.Name)
}

func (c *Conversion) reverse() *Conversion {
	return &Conversion{
		Root:   c.Root,
		To:     !c.To,
		Name:   c.Name,
		Source: c.Target,
		Target: c.Source,
	}
}

func (c *Conversion) subConversions() ([]*Conversion, error) {
	conversions := []*Conversion{}
	for _, field := range c.Target.Fields {
		if field.Kind == Struct {
			log.Printf("find pair for struct field %v", field)
			_, source, err := findSourceField(c.Source.Fields, field)
			log.Printf("field %v <- source %v to=%v", field, source, c.To)
			if err != nil {
				return nil, fmt.Errorf("cannot find pair for struct %v at %v", field, c.Source.Fields)
			}
			subConversion := &Conversion{
				Root:   false,
				To:     c.To,
				Name:   c.Name,
				Source: source,
				Target: field,
			}
			conversions = append(conversions, subConversion)
		}
	}
	return conversions, nil
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
		if tl.WrappedType.External == nil || tl.WrappedType.Internal == nil {
			return "", fmt.Errorf("both wrapped internal and external types muts be set")
		}
		tl.defineWrapperType(f)
		for _, wm := range tl.WrappedType.WrapperMethods {
			tl.implementAPICallWrapping(f, wm)
			f.Empty()
		}
		if err := tl.writeConversions(f); err != nil {
			return "", fmt.Errorf("failed to write conversion functions: %w", err)
		}
	}
	return f.GoString(), nil
}

func (tl *TranslationLayer) writeConversions(f *jen.File) error {
	conversions := []*Conversion{
		{
			Root:   true,
			To:     true,
			Name:   tl.WrappedType.ExternalName,
			Source: tl.WrappedType.Internal,
			Target: tl.WrappedType.External,
		},
	}
	for i := 0; i < len(conversions); i++ {
		conversion := conversions[i]
		log.Printf("Converting %v %d conversions left", conversion, len(conversions))
		subConversions, err := conversion.subConversions()
		if err != nil {
			return fmt.Errorf("failed to extract sub conversion conversion list: %w", err)
		}
		conversions = append(conversions, subConversions...)
		log.Printf("Added %d sub-conversions, %d conversions left", len(subConversions), len(conversions))

		if err := writeConversionFunc(f, conversion); err != nil {
			return fmt.Errorf("failed to generate conversion to external type: %w", err)
		}
		f.Empty()
		conversion = conversion.reverse()
		if err := writeConversionFunc(f, conversion); err != nil {
			return fmt.Errorf("failed to generate conversion to external type: %w", err)
		}
		f.Empty()
	}
	log.Print("Done with conversions")
	return nil
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

func writeConversionFunc(f *jen.File, conversion *Conversion) error {
	source := conversion.Source
	source.NamedType.Name = source.Alias
	conversionBody, err := returnConversion(
		conversion.Target,
		conversion.Source,
		conversion.direction(),
		conversion.Name,
	)
	if err != nil {
		return fmt.Errorf("struct conversion failed: %v", err)
	}
	addFunctionSignature(
		f,
		&FunctionSignature{
			Name:    conversion.method(),
			Args:    []NamedType{source.NamedType},
			Returns: []NamedType{conversion.Target.NamedType},
		},
		conversionBody,
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

func returnConversion(dst, src *ComplexType, direction, externalName string) (*jen.Statement, error) {
	typeConversion, err := convertTypes(dst, src, direction, externalName)
	if err != nil {
		return nil, err
	}
	return jen.Return(typeConversion), nil
}

func convertTypes(dst, src *ComplexType, direction, externalName string) (*jen.Statement, error) {
	fieldConversions, err := convertFields(dst, src, direction, externalName)
	if err != nil {
		return nil, err
	}
	if dst.isPointer() {
		return jen.Op("&").Id(string(dst.Type.dereference())).Values(fieldConversions), nil
	}
	return jen.Id(string(dst.Type.String())).Values(fieldConversions), nil
}

func convertFields(dst, src *ComplexType, direction, externalName string) (jen.Dict, error) {
	remaining := src.Fields
	values := jen.Dict{}
	for _, field := range dst.Fields {
		var err error
		var srcField *ComplexType
		var conversion *jen.Statement
		remaining, srcField, err = findSourceField(remaining, field)
		if err != nil {
			return nil, fmt.Errorf("failed to match conversion pair: %w", err)
		}
		if field.Kind != SimpleField {
			cm := conversionMethod(field.Type, direction, externalName)
			values[jen.Id(field.Name)] = jen.Id(cm).Call(
				jen.Id(src.Name).Dot(srcField.Name),
			)
			continue
		}
		conversion, err = computeAssignment(field, src, srcField.NamedType)
		if err != nil {
			return nil, fmt.Errorf("failed to compute conversion: %w", err)
		}
		values[jen.Id(field.Name)] = conversion
	}
	return values, nil
}

func findSourceField(candidates []*ComplexType, target *ComplexType) ([]*ComplexType, *ComplexType, error) {
	prefix := []*ComplexType{}
	for i, candidate := range candidates {
		if strings.EqualFold(target.Name, candidate.Name) && target.assignableFrom(candidate.NamedType) {
			remaining := append(prefix, candidates[i+1:]...)
			if reflect.DeepEqual(remaining, candidates) {
				panic(fmt.Sprintf("remaining cannot match candidates, source %v has not been extracted from %v",
					candidate, remaining))
			}
			return remaining, candidate, nil
		}
		prefix = append(prefix, candidate)
	}
	return nil, nil, fmt.Errorf("could not find corresponding field for %v at %v", target, candidates)
}

func computeAssignment(target *ComplexType, src *ComplexType, srcField NamedType) (*jen.Statement, error) {
	switch {

	// Same types on both sides
	case srcField.Type == target.Type:
		return jen.Id(src.Name).Dot(srcField.Name), nil

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

	// Target can be converted from the primitive type of the source field
	case target.Primitive != nil && *target.Primitive == srcField.Type.dereference():
		primitive, ok := target.primitive()
		if !ok {
			return nil, fmt.Errorf("could not cast target %v to primitive type", target)
		}
		assignment, err := computeAssignment(primitive, src, srcField)
		if err != nil {
			return nil, fmt.Errorf("failed to assign after casting field to primitive: %w", err)
		}
		return jen.Id(target.Type.dereference().String()).Call(assignment), nil

	// Source field can be converted to the primitive type of the target
	case srcField.Primitive != nil && *srcField.Primitive == target.Type.dereference():
		primitive, ok := srcField.primitive()
		if !ok {
			return nil, fmt.Errorf("could not cast field %v.%v to primitive type", src, srcField)
		}
		assignment, err := computeAssignment(target, src, primitive)
		if err != nil {
			return nil, fmt.Errorf("failed to assign after casting field to primitive: %w", err)
		}
		return jen.Id(primitive.Type.dereference().String()).Call(assignment), nil

	default:
		return nil, fmt.Errorf("cannot find way to assign %s.%v to %v", src, srcField, target)
	}
}

func conversionMethod(t Type, direction string, externalName string) string {
	return fmt.Sprintf("%s%s%s",
		firstToLower(t.dereference().base()), strings.Title(direction), externalName)
}

func firstToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}

func firstToUpper(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToUpper(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}
