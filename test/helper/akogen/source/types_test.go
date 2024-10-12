package source_test

import (
	"testing"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/metadata"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindAnnotatedType(t *testing.T) {
	sg, err := source.New("../sample/def.go")
	require.NoError(t, err)
	ts := sg.FindAnnotatedType("+akogen:InternalType")
	require.NotNil(t, ts)
	assert.Equal(t, "Resource", ts.Name.Name)
}

func TestDescribeType(t *testing.T) {
	sg, err := source.New("../sample/def.go")
	require.NoError(t, err)
	expected := metadata.NewStruct(
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
	dt, err := sg.DescribeType(
		"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/sample",
		"Resource",
	)
	require.NoError(t, err)
	assert.Equal(t, expected, dt)
}
