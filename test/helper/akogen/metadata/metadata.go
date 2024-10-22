package metadata

import (
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Import struct {
	Alias, Path string
}

type FunctionSignature struct {
	Name    string
	Args    NamedTypes
	Returns NamedTypes
}

type MethodSignature struct {
	FunctionSignature
	Receiver NamedType
}

type Interface struct {
	Name       string
	Operations []FunctionSignature
}

func removeBase(s, pkgPath string) string {
	if len(s) == 0 {
		return ""
	}
	base := filepath.Base(pkgPath)
	if strings.HasPrefix(s, base) && len(s) > len(base) {
		return firstToLower(strings.Replace(s, base, "", 1))
	}
	return s
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
