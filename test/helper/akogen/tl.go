package akogen

import (
	"errors"
	"fmt"

	"github.com/dave/jennifer/jen"
)

var (
	// ErrorNilSpec code spec cannot be nil
	ErrNilSpec = errors.New("code specification cannot be nil")

	// ErrorEmptySpec code spec cannot be empty
	ErrEmptySpec = errors.New("code specification cannot be empty")
)

type TranslationLayer struct {
	PackageName string
	WrappedType *WrappedType
}

func (tl *TranslationLayer) Generate() (string, error) {
	if tl == nil {
		return "", ErrNilSpec
	}
	if tl.isEmpty() {
		return "", ErrEmptySpec
	}
	f := jen.NewFile(tl.PackageName)
	if tl.WrappedType != nil {
		if err := tl.WrappedType.generate(f); err != nil {
			return "", fmt.Errorf("failed to generate wrapper type: %w", err)
		}
		conversion := NewConversion(
			tl.WrappedType.ExternalName,
			tl.WrappedType.External,
			tl.WrappedType.Internal,
		)
		if err := conversion.generate(f); err != nil {
			return "", fmt.Errorf("failed to generate conversion functions: %w", err)
		}
	}
	return f.GoString(), nil
}

func (tl *TranslationLayer) isEmpty() bool {
	return tl.PackageName == "" && tl.WrappedType == nil
}
