package metadata

import (
	"fmt"
	"strings"
)

type Type string

func (t Type) String() string {
	return string(t)
}

func (t Type) IsPointer() bool {
	return t.String()[0] == '*'
}

func (t Type) Dereference() Type {
	for t.IsPointer() {
		return Type(t.String()[1:])
	}
	return t
}

func (t Type) StripPackage(pkgPath string) Type {
	stripped := strings.Replace(string(t), fmt.Sprintf("%s.", pkgPath), "", 1)
	return Type(stripped)
}

func (t Type) Base() string {
	parts := strings.Split(string(t.Dereference()), ".")
	return parts[len(parts)-1]
}

func (t Type) Lib() string {
	deref := t.Dereference()
	totalSize := len(string(deref))
	baseSize := len(t.Base())
	if baseSize == totalSize {
		return ""
	}
	// return all but the dot and base suffix
	return string(deref)[:totalSize-baseSize-1]
}

func (t Type) Pointer() Type {
	return Type(fmt.Sprintf("*%v", t))
}
