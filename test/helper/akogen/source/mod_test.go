package source_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/source"
)

func TestModPathTest(t *testing.T) {
	modPath, _, err := source.ModPath("../source/mod_test.go")
	require.NoError(t, err)
	assert.Equal(t, "github.com/mongodb/mongodb-atlas-kubernetes/v2", modPath)
}

func TestPkgPathTest(t *testing.T) {
	pkgPath, err := source.PkgPathFor("mod_test.go")
	require.NoError(t, err)
	assert.Equal(t, "github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/source", pkgPath)
}

func TestFilePathFor(t *testing.T) {
	for _, tc := range []struct {
		title       string
		packageName string
		want        string
	}{
		{
			title:       "a go standard lib package renders expected file path",
			packageName: "fmt",
			want:        expectStdLibPath("fmt"),
		},
		{
			title:       "3rd party package renders expected file path",
			packageName: "github.com/stretchr/testify/require",
			want:        expect3rdPartyModPath("pkg/mod/github.com/stretchr/testify", "v1.9.0"),
		},
		{
			title:       "a package in this mod renders expected file path",
			packageName: "github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib",
			want:        expectCurrentModPath(t, "test/helper/akogen/lib"),
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			dir, err := source.FilePathFor(tc.packageName)
			require.NoError(t, err)
			assert.Equal(t, tc.want, dir)
		})
	}
}

func expectStdLibPath(pkgName string) string {
	return filepath.Join(source.GoRoot(), "src", pkgName)
}

func expectCurrentModPath(t *testing.T, path string) string {
	_, baseDir, err := source.ModPath(".")
	require.NoError(t, err)
	return filepath.Join(baseDir, path)
}

func expect3rdPartyModPath(dir, version string) string {
	return filepath.Join(source.GoPath(), fmt.Sprintf("%s@%s", dir, version))
}
