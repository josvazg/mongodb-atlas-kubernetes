package akogen

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Type string

func (t Type) String() string {
	return string(t)
}

func (t Type) isPointer() bool {
	return t.String()[0] == '*'
}

func (t Type) dereference() Type {
	for t.isPointer() {
		return Type(t.String()[1:])
	}
	return t
}

func (t Type) base() string {
	parts := strings.Split(string(t), ".")
	return parts[len(parts)-1]
}

func (t Type) pointer() Type {
	return Type(fmt.Sprintf("*%v", t))
}

func (t Type) generateZeroValue() *jen.Statement {
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
	if strings.Contains("bool", string(t)) {
		return jen.False()
	}
	// TODO be smarter about non primitive types?
	return jen.Nil()
}

