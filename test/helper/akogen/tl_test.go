package akogen_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/sample"
)

const (
	wrappedCallSample = `// Generated by AKOGen code Generator - do not edit

package %s

import (
	"context"
	pointer "github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/pointer"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib"
)

type wrapper struct {
	api lib.API
}

func (w *wrapper) Create(ctx context.Context, r *Resource) (*Resource, error) {
	libR, err := w.api.Create(ctx, toAtlas(r))
	if err != nil {
		return nil, err
	}
	return fromAtlas(libR), nil
}

func (w *wrapper) Get(ctx context.Context, s string) (*Resource, error) {
	libR, err := w.api.Get(ctx, s)
	if err != nil {
		return nil, err
	}
	return fromAtlas(libR), nil
}

func toAtlas(r *Resource) *lib.Resource {
	if r == nil {
		return nil
	}
	return &lib.Resource{
		ComplexSubtype: complexSubtypeToAtlas(r.ComplexSubtype),
		Enabled:        pointer.MakePtr(r.Enabled),
		Id:             r.ID,
		OptionalRef:    optionalRefToAtlas(r.OptionalRef),
		SelectedOption: pointer.MakePtr(string(r.SelectedOption)),
		Status:         pointer.MakePtr(r.Status),
	}
}

func fromAtlas(libR *lib.Resource) *Resource {
	if libR == nil {
		return nil
	}
	return &Resource{
		ComplexSubtype: complexSubtypeFromAtlas(libR.ComplexSubtype),
		Enabled:        pointer.GetOrDefault(libR.Enabled, false),
		ID:             libR.Id,
		OptionalRef:    optionalRefFromAtlas(libR.OptionalRef),
		SelectedOption: OptionType(pointer.GetOrDefault(libR.SelectedOption, "")),
		Status:         pointer.GetOrDefault(libR.Status, ""),
	}
}

func complexSubtypeToAtlas(cs ComplexSubtype) lib.ComplexSubtype {
	return lib.ComplexSubtype{
		Name:    cs.Name,
		Subtype: string(cs.Subtype),
	}
}

func complexSubtypeFromAtlas(libCs lib.ComplexSubtype) ComplexSubtype {
	return ComplexSubtype{
		Name:    libCs.Name,
		Subtype: Subtype(libCs.Subtype),
	}
}

func optionalRefToAtlas(or *OptionalRef) *lib.OptionalRef {
	if or == nil {
		return nil
	}
	return &lib.OptionalRef{Ref: or.Ref}
}

func optionalRefFromAtlas(libOr *lib.OptionalRef) *OptionalRef {
	if libOr == nil {
		return nil
	}
	return &OptionalRef{Ref: libOr.Ref}
}

// Generated by AKOGen code Generator - do not edit
`
)

func fullSample(packageName string) *akogen.TranslationLayer {
	return &akogen.TranslationLayer{
		PackageName: packageName,
		WrappedType: &akogen.WrappedType{
			Translation: akogen.Translation{
				Lib: akogen.Import{
					Alias: "lib",
					Path:  "github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib",
				},
				ExternalName: "Atlas",
				External: akogen.NewStruct(
					akogen.NewNamedType("libR", "*github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib.Resource"),
					akogen.NewStructField(
						"ComplexSubtype",
						akogen.NewNamedType("libCs", "github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib.ComplexSubtype"),
						akogen.NewSimpleField("Name", "string"),
						akogen.NewSimpleField("Subtype", "string"),
					),
					akogen.NewSimpleField("Enabled", "*bool"),
					akogen.NewSimpleField("Id", "string"),
					akogen.NewStructField(
						"OptionalRef",
						akogen.NewNamedType("libOr", "*github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib.OptionalRef"),
						akogen.NewSimpleField("Ref", "string"),
					),
					akogen.NewSimpleField("SelectedOption", "*string"),
					akogen.NewSimpleField("Status", "*string"),
				),
				ExternalAPI: akogen.NewNamedType("api", "API"),
				Internal: akogen.NewStruct(
					akogen.NewNamedType("r", "*Resource"),
					akogen.NewStructField(
						"ComplexSubtype",
						akogen.NewNamedType("cs", "ComplexSubtype"),
						akogen.NewSimpleField("Name", "string"),
						akogen.NewSimpleField("Subtype", "Subtype").WithPrimitive("string"),
					),
					akogen.NewSimpleField("Enabled", "bool"),
					akogen.NewSimpleField("ID", "string"),
					akogen.NewStructField(
						"OptionalRef",
						akogen.NewNamedType("or", "*OptionalRef"),
						akogen.NewSimpleField("Ref", "string"),
					),
					akogen.NewSimpleField("SelectedOption", "OptionType").WithPrimitive("string"),
					akogen.NewSimpleField("Status", "string"),
				),
				Wrapper: akogen.NewNamedType("w", "wrapper"),
			},
			WrapperMethods: []akogen.WrapperMethod{
				{
					MethodSignature: akogen.MethodSignature{
						Receiver: akogen.NewNamedType("w", "*wrapper"),
						FunctionSignature: akogen.FunctionSignature{
							Name: "Create",
							Args: []akogen.NamedType{
								akogen.NewNamedType("ctx", "context.Context"),
								akogen.NewNamedType("r", "*Resource"),
							},
							Returns: []akogen.NamedType{
								akogen.NewNamedType("r", "*Resource"),
								akogen.NewNamedType("err", "error"),
							},
						},
					},
					WrappedCall: akogen.FunctionSignature{
						Name: "Create",
						Args: []akogen.NamedType{
							akogen.NewNamedType("ctx", "context.Context"),
							akogen.NewNamedType("libR", "*github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib.Resource"),
						},
						Returns: []akogen.NamedType{
							akogen.NewNamedType("libR", "*github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib.Resource"),
							akogen.NewNamedType("err", "error"),
						},
					},
				},
				{
					MethodSignature: akogen.MethodSignature{
						Receiver: akogen.NewNamedType("w", "*wrapper"),
						FunctionSignature: akogen.FunctionSignature{
							Name: "Get",
							Args: akogen.NamedTypes{
								akogen.NewNamedType("ctx", "context.Context"),
								akogen.NewNamedType("s", "string"),
							},
							Returns: akogen.NamedTypes{
								akogen.NewNamedType("r", "*Resource"),
								akogen.NewNamedType("err", "error"),
							},
						},
					},
					WrappedCall: akogen.FunctionSignature{
						Name: "Get",
						Args: []akogen.NamedType{
							akogen.NewNamedType("ctx", "context.Context"),
							akogen.NewNamedType("s", "string"),
						},
						Returns: []akogen.NamedType{
							akogen.NewNamedType("libR", "*github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib.Resource"),
							akogen.NewNamedType("err", "error"),
						},
					},
				},
			},
		},
	}
}

func TestNewTranslationLayer(t *testing.T) {
	packageName := "sample"
	defaults := akogen.DefaultSettings
	defaults.ExternalName = "Atlas"
	defaults.WrapperType = "wrapper"

	got := akogen.NewTranslationLayer(&akogen.TranslationLayerSpec{
		PackageName:  packageName,
		Name:         "Resource",
		API:          reflect.TypeOf((*lib.API)(nil)).Elem(),
		ExternalType: &lib.Resource{},
		InternalType: &sample.Resource{},
	}, defaults)
	want := fullSample(packageName)
	assert.Equal(t, want, got)
}

func TestGenerateTranslationLayer(t *testing.T) {
	packageName := randomString(t, "prefix")
	for _, tc := range []struct {
		title     string
		spec      *akogen.TranslationLayer
		want      string
		wantError error
	}{
		{
			title:     "nil spec fails",
			wantError: akogen.ErrNilSpec,
		},
		{
			title:     "empty spec fails",
			spec:      &akogen.TranslationLayer{},
			wantError: akogen.ErrEmptySpec,
		},
		{
			title: "setting just the package generates source of such package",
			spec: &akogen.TranslationLayer{
				PackageName: packageName,
			},
			want: fmt.Sprintf("// Generated by AKOGen code Generator - do not edit\n\npackage %s\n", packageName),
		},
		{
			title: "specifying a full sample wrapper generates the expected wrapper code",
			spec:  fullSample(packageName),
			want:  fmt.Sprintf(wrappedCallSample, packageName),
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			src, err := tc.spec.Generate()
			if tc.wantError != nil {
				require.Empty(t, src)
				assert.Equal(t, tc.wantError, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, src)
		})
	}
}

func TestTranslationLayerGenerateFile(t *testing.T) {
	packageName := "sample"
	tl := fullSample(packageName)
	srcCode, err := tl.Generate()
	require.NoError(t, err)
	err = os.WriteFile("sample/generated.go", ([]byte)(srcCode), 0600)
	require.NoError(t, err)
}

func randomString(t *testing.T, prefix string) string {
	n, err := rand.Int(rand.Reader, big.NewInt(99999))
	require.NoError(t, err)
	return fmt.Sprintf("%s%d", prefix, n.Int64())
}