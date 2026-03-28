package pdf

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFMultiPage(t, path, 3)

	count, err := PageCount(path)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestPageCount_FileNotFound(t *testing.T) {
	_, err := PageCount("/nonexistent/path.pdf")
	require.Error(t, err)
}

func TestExtractPageText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFWithText(t, path, "Hello World")

	text, err := ExtractPageText(path, 1)
	require.NoError(t, err)
	assert.Contains(t, text, "Hello World")
}

func TestExtractPageText_InvalidPage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFMultiPage(t, path, 1)

	_, err := ExtractPageText(path, 99)
	require.Error(t, err)
}

func TestExtractPageText_SeparateWords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFSeparateWords(t, path)

	text, err := ExtractPageText(path, 1)
	require.NoError(t, err)
	assert.Contains(t, text, "Hello")
	assert.Contains(t, text, "World")
	assert.Contains(t, text, "Second")
}

func TestSearchText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFWithText(t, path, "Hello World")

	results, err := SearchText(path, "Hello")
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Equal(t, 1, results[0].Page)
	assert.Contains(t, results[0].Snippet, "Hello")
}

func TestSearchText_NoMatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pdf")
	createTestPDFWithText(t, path, "Hello World")

	results, err := SearchText(path, "nonexistent_xyz_999")
	require.NoError(t, err)
	assert.Empty(t, results)
}
