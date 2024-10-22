package source_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/source"
)

func TestGetAnnotations(t *testing.T) {
	expected := []source.GenAnnotation{
		{Raw: "// +akogen:ExternalSystem:Atlas", Name: "ExternalSystem", Type: source.SimpleValue, Value: "Atlas"},
		{Raw: `// +akogen:ExternalPackage:var=lib,path="github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib"`,
			Name: "ExternalPackage", Type: source.ArgsValues,
			Args: map[string]string{"var": "lib", "path": "github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib"}},
		{Raw: `// +akogen:ExternalType:var=res,type=*lib.Resource`, Name: "ExternalType", Type: source.ArgsValues,
			Args: map[string]string{"var": "res", "type": "*lib.Resource"}},
		{Raw: `// +akogen:ExternalAPI:var=api,type=lib.API`, Name: "ExternalAPI", Type: source.ArgsValues,
			Args: map[string]string{"var": "api", "type": "lib.API"}},
		{Raw: `// +akogen:WrapperType:var="w",type="Wrapper"`, Name: "WrapperType", Type: source.ArgsValues,
			Args: map[string]string{"var": "w", "type": "Wrapper"}},
		{Raw: `// +akogen:InternalType:var=res,pointer=true`, Name: "InternalType", Type: source.ArgsValues,
			Args: map[string]string{"var": "res", "pointer": "true"}},
	}
	sg, err := source.NewFromFile("../sample/def.go")
	require.NoError(t, err)
	annotations, err := sg.GenAnnotationsFor("akogen")
	require.NoError(t, err)
	assert.Equal(t, expected, annotations)
}
