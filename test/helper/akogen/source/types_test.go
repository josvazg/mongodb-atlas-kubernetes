package source_test

import (
	"testing"

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
