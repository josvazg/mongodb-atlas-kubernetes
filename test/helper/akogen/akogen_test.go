package akogen_test

import (
	"fmt"
	"math/rand"
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
	packageName := randomString("prefix")
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
						External:     akogen.NamedType{Name: "apiRes", Type: "*api.Resource"},
						ExternalAPI:  akogen.NamedType{Name: "api", Type: "API"},
						Internal:     akogen.NamedType{Name: "res", Type: "*Resource"},
						Wrapper:      akogen.NamedType{Name: "w", Type: "wrapper"},
					},
					//NamedType: akogen.NamedType{Name: "api", Type: "API"},
					WrapperMethods: []akogen.WrapperMethod{
						{
							MethodSignature: akogen.MethodSignature{
								ImplType: akogen.NamedType{Name: "w", Type: "*wrapper"},
								FunctionSignature: akogen.FunctionSignature{
									Name: "Get",
									Args: []akogen.NamedType{
										{Name: "ctx", Type: "context.Context"},
										{Name: "id", Type: "string"},
									},
									Returns: []akogen.NamedType{
										{Name: "res", Type: "*Resource"},
										{Name: "err", Type: "error"},
									},
								},
							},
							WrappedCall: akogen.FunctionSignature{
								Name: "Get",
								Args: []akogen.NamedType{
									{Name: "ctx", Type: "context.Context"},
									{Name: "id", Type: "string"},
								},
								Returns: []akogen.NamedType{
									{Name: "res", Type: "*api.Resource"},
									{Name: "err", Type: "error"},
								},
							},
						},

						{
							MethodSignature: akogen.MethodSignature{
								ImplType: akogen.NamedType{Name: "w", Type: "*wrapper"},
								FunctionSignature: akogen.FunctionSignature{
									Name: "Create",
									Args: []akogen.NamedType{
										{Name: "ctx", Type: "context.Context"},
										{Name: "res", Type: "*Resource"},
									},
									Returns: []akogen.NamedType{
										{Name: "res", Type: "*Resource"},
										{Name: "err", Type: "error"},
									},
								},
							},
							WrappedCall: akogen.FunctionSignature{
								Name: "Create",
								Args: []akogen.NamedType{
									{Name: "ctx", Type: "context.Context"},
									{Name: "apiRes", Type: "*api.Resource"},
								},
								Returns: []akogen.NamedType{
									{Name: "res", Type: "*api.Resource"},
									{Name: "err", Type: "error"},
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

func randomString(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, rand.Intn(99999))
}
