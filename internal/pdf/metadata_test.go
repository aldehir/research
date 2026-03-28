package pdf

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractMetadata_PageCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFMultiPage(t, path, 3)

	meta, err := ExtractMetadata(path)
	require.NoError(t, err)
	assert.Equal(t, 3, meta.PageCount)
}

func TestExtractMetadata_FileNotFound(t *testing.T) {
	_, err := ExtractMetadata("/nonexistent/path.pdf")
	require.Error(t, err)
}
