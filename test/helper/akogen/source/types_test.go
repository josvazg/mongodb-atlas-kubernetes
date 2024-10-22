package source_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/metadata"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/source"
)

func TestFindAnnotatedType(t *testing.T) {
	sg, err := source.NewFromFile("../sample/def.go")
	require.NoError(t, err)
	ts := sg.FindAnnotatedType("+akogen:InternalType")
	require.NotNil(t, ts)
	assert.Equal(t, "Resource", ts.Name.Name)
}

func TestDescribeType(t *testing.T) {
	sg, err := source.NewFromFile("../sample/def.go")
	require.NoError(t, err)
	want := metadata.NewStruct(
		metadata.NewNamedType("Resource", "Resource"),
		metadata.NewStructField(
			"ComplexSubtype",
			metadata.NewNamedType("ComplexSubtype", "ComplexSubtype"),
			metadata.NewSimpleField("Name", "string"),
			metadata.NewSimpleField("Subtype", "Subtype").WithPrimitive("string"),
		),
		metadata.NewSimpleField("Enabled", "bool"),
		metadata.NewSimpleField("ID", "string"),
		metadata.NewStructField(
			"OptionalRef",
			metadata.NewNamedType("OptionalRef", "*OptionalRef"),
			metadata.NewSimpleField("Ref", "string"),
		),
		metadata.NewSimpleField("SelectedOption", "OptionType").WithPrimitive("string"),
		metadata.NewSimpleField("Status", "string"),
	)
	dt, err := sg.DescribeType("Resource")
	require.NoError(t, err)
	assert.Equal(t, want, dt)
}

func TestDescribeInterface(t *testing.T) {
	sg, err := source.NewFromPackage("github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib")
	require.NoError(t, err)
	want := &metadata.Interface{
		Name: "API",
		Operations: []metadata.FunctionSignature{
			{
				Name: "Create",
				Args: []metadata.NamedType{
					metadata.NewNamedType("ctx", "context.Context"),
					metadata.NewNamedType("apiRes", "*Resource"),
				},
				Returns: []metadata.NamedType{
					metadata.NewNamedType("", "*Resource"),
					metadata.NewNamedType("", "error"),
				},
			},
			{
				Name: "Get",
				Args: []metadata.NamedType{
					metadata.NewNamedType("ctx", "context.Context"),
					metadata.NewNamedType("id", "string"),
				},
				Returns: []metadata.NamedType{
					metadata.NewNamedType("", "*Resource"),
					metadata.NewNamedType("", "error"),
				},
			},
		},
	}
	di, err := sg.DescribeInterface("API")
	require.NoError(t, err)
	assert.Equal(t, want, di)
}
