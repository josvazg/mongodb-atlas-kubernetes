package akogen_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen"
)

const (
	wrappedCallSample = `package %s

import "some/path/to/lib"

type wrapper struct {
	api lib.API
}

func (w *wrapper) Get(ctx context.Context, id string) (*Resource, error) {
	res, err := w.api.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return fromAtlas(res), nil
}

func (w *wrapper) Create(ctx context.Context, res *Resource) (*Resource, error) {
	res, err := w.api.Create(ctx, toAtlas(res))
	if err != nil {
		return nil, err
	}
	return fromAtlas(res), nil
}
`
)

func TestGenAPIWrapper(t *testing.T) {
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
			want: fmt.Sprintf("package %s\n", packageName),
		},
		{
			title: "specifying a wrapper generates the expected wrapper code",
			spec: &akogen.TranslationLayer{
				PackageName: packageName,
				WrappedType: &akogen.WrappedType{
					Translation: akogen.Translation{
						Lib:          akogen.Import{Alias: "lib", Path: "some/path/to/lib"},
						ExternalName: "Atlas",
						External:     akogen.Struct{NamedType: akogen.NewNamedType("apiRes", "*api.Resource")},
						ExternalAPI:  akogen.NewNamedType("api", "API"),
						Internal:     akogen.Struct{NamedType: akogen.NewNamedType("res", "*Resource")},
						Wrapper:      akogen.NewNamedType("w", "wrapper"),
					},
					WrapperMethods: []akogen.WrapperMethod{
						{
							MethodSignature: akogen.MethodSignature{
								Receiver: akogen.NewNamedType("w", "*wrapper"),
								FunctionSignature: akogen.FunctionSignature{
									Name: "Get",
									Args: akogen.NamedTypes{
										akogen.NewNamedType("ctx", "context.Context"),
										akogen.NewNamedType("id", "string"),
									},
									Returns: akogen.NamedTypes{
										akogen.NewNamedType("res", "*Resource"),
										akogen.NewNamedType("err", "error"),
									},
								},
							},
							WrappedCall: akogen.FunctionSignature{
								Name: "Get",
								Args: []akogen.NamedType{
									akogen.NewNamedType("ctx", "context.Context"),
									akogen.NewNamedType("id", "string"),
								},
								Returns: []akogen.NamedType{
									akogen.NewNamedType("res", "*api.Resource"),
									akogen.NewNamedType("err", "error"),
								},
							},
						},
						{
							MethodSignature: akogen.MethodSignature{
								Receiver: akogen.NewNamedType("w", "*wrapper"),
								FunctionSignature: akogen.FunctionSignature{
									Name: "Create",
									Args: []akogen.NamedType{
										akogen.NewNamedType("ctx", "context.Context"),
										akogen.NewNamedType("res", "*Resource"),
									},
									Returns: []akogen.NamedType{
										akogen.NewNamedType("res", "*Resource"),
										akogen.NewNamedType("err", "error"),
									},
								},
							},
							WrappedCall: akogen.FunctionSignature{
								Name: "Create",
								Args: []akogen.NamedType{
									akogen.NewNamedType("ctx", "context.Context"),
									akogen.NewNamedType("apiRes", "*api.Resource"),
								},
								Returns: []akogen.NamedType{
									akogen.NewNamedType("res", "*api.Resource"),
									akogen.NewNamedType("err", "error"),
								},
							},
						},
					},
				},
			},
			want: fmt.Sprintf(wrappedCallSample, packageName),
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			src, err := akogen.GenerateTranslationLayer(tc.spec)
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

func randomString(t *testing.T, prefix string) string {
	n, err := rand.Int(rand.Reader, big.NewInt(99999))
	require.NoError(t, err)
	return fmt.Sprintf("%s%d", prefix, n.Int64())
}
