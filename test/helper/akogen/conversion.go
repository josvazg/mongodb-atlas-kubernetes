package akogen

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dave/jennifer/jen"
)

const (
	pointerLib = "github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/pointer"
)

type Conversion struct {
	Root   bool
	To     bool
	Name   string
	Source *ComplexType
	Target *ComplexType
}

func NewConversion(name string, target, source *ComplexType) *Conversion {
	return &Conversion{
		Root:   true,
		To:     true,
		Name:   name,
		Source: source,
		Target: target,
	}
}

func (c *Conversion) String() string {
	return fmt.Sprintf("{Method: %v, Source: %v, Target: %v", c.method(), c.Source, c.Target)
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

func (c *Conversion) generate(f *jen.File) error {
	conversions := []*Conversion{c}
	for i := 0; i < len(conversions); i++ {
		conversion := conversions[i]
		subConversions, err := conversion.subConversions()
		if err != nil {
			return fmt.Errorf("failed to extract sub conversion conversion list: %w", err)
		}
		conversions = append(conversions, subConversions...)

		if err := conversion.generateFunc(f); err != nil {
			return fmt.Errorf("failed to generate conversion to external type: %w", err)
		}
		f.Empty()
		conversion = conversion.reverse()
		if err := conversion.generateFunc(f); err != nil {
			return fmt.Errorf("failed to generate conversion to external type: %w", err)
		}
		f.Empty()
	}
	return nil
}

func (c *Conversion) generateFunc(f *jen.File) error {
	source := c.Source
	source.NamedType.Name = source.Alias
	returnCode, err := c.generateReturn()
	if err != nil {
		return fmt.Errorf("struct conversion failed: %v", err)
	}
	generateFunctionSignature(
		f,
		&FunctionSignature{
			Name:    c.method(),
			Args:    []NamedType{source.NamedType},
			Returns: []NamedType{c.Target.NamedType},
		},
		generateReturnNilOnNilArg(c.Source.NamedType),
		returnCode,
	)
	return nil
}

func (c *Conversion) subConversions() ([]*Conversion, error) {
	conversions := []*Conversion{}
	for _, field := range c.Target.Fields {
		if field.Kind == Struct {
			_, source, err := findSourceField(c.Source.Fields, field)
			if err != nil {
				return nil, fmt.Errorf("cannot find pair for struct %v at %v", field, c.Source.Fields)
			}
			subConversion := c.subConversion(field, source)
			conversions = append(conversions, subConversion)
		}
	}
	return conversions, nil
}

func (c *Conversion) subConversion(target, source *ComplexType) *Conversion {
	return &Conversion{
		Root:   false,
		To:     c.To,
		Name:   c.Name,
		Source: source,
		Target: target,
	}
}

func (c *Conversion) generateReturn() (*jen.Statement, error) {
	typeConversion, err := c.generateAssignmentExpression()
	if err != nil {
		return nil, err
	}
	return jen.Return(typeConversion), nil
}

func (c *Conversion) generateAssignmentExpression() (*jen.Statement, error) {
	fieldConversions, err := c.generateFields()
	if err != nil {
		return nil, err
	}
	if c.Target.isPointer() {
		return jen.Op("&").Id(string(c.Target.Type.dereference())).Values(fieldConversions), nil
	}
	return jen.Id(string(c.Target.Type.String())).Values(fieldConversions), nil
}

func (c *Conversion) generateFields() (jen.Dict, error) {
	remaining := c.Source.Fields
	values := jen.Dict{}
	for _, field := range c.Target.Fields {
		var err error
		var srcField *ComplexType
		var conversion *jen.Statement
		remaining, srcField, err = findSourceField(remaining, field)
		if err != nil {
			return nil, fmt.Errorf("failed to match conversion pair: %w", err)
		}
		if field.Kind != SimpleField {
			subConversion := c.subConversion(field, srcField)
			cm := subConversion.method()
			values[jen.Id(field.Name)] = jen.Id(cm).Call(
				jen.Id(c.Source.Name).Dot(srcField.Name),
			)
			continue
		}
		conversion, err = generateAssignment(field, c.Source, srcField.NamedType)
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

func generateAssignment(target *ComplexType, src *ComplexType, srcField NamedType) (*jen.Statement, error) {
	switch {

	// Same types on both sides
	case srcField.Type == target.Type:
		return jen.Id(src.Name).Dot(srcField.Name), nil

	// Target is a pointer
	case target.isPointer() && !srcField.isPointer():
		assignment, err := generateAssignment(target, src, srcField.pointer())
		if err != nil {
			return nil, fmt.Errorf("failed to assign to pointer target %v: %w", target, err)
		}
		return jen.Qual(pointerLib, "MakePtr").Call(assignment), nil

	// Source field is the pointer
	case srcField.isPointer() && !target.isPointer():
		deref := srcField.dereference()
		assignment, err := generateAssignment(target, src, deref)
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
		assignment, err := generateAssignment(primitive, src, srcField)
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
		assignment, err := generateAssignment(target, src, primitive)
		if err != nil {
			return nil, fmt.Errorf("failed to assign after casting field to primitive: %w", err)
		}
		return jen.Id(primitive.Type.dereference().String()).Call(assignment), nil

	default:
		return nil, fmt.Errorf("cannot find way to assign %s.%v to %v", src, srcField, target)
	}
}

func generateReturnNilOnNilArg(nt NamedType) *jen.Statement {
	if nt.isPointer() {
		return jen.If(jen.Id(nt.Name).Op("==").Nil()).Block(
			jen.Return(jen.Nil()),
		)
	}
	return nil
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
