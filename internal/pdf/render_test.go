package pdf

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPage(t *testing.T) {
	t.Run("renders a page to PNG bytes", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.pdf")
		createTestPDFWithText(t, path, "Hello World")

		png, err := RenderPage(path, 1)
		require.NoError(t, err)
		require.NotEmpty(t, png)
		// PNG files start with the 8-byte magic header
		assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, png[:8])
	})

	t.Run("error on invalid page number", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.pdf")
		createTestPDFWithText(t, path, "Hello World")

		_, err := RenderPage(path, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")

		_, err = RenderPage(path, 99)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "out of range")
	})

	t.Run("error on nonexistent file", func(t *testing.T) {
		_, err := RenderPage("/nonexistent/path.pdf", 1)
		require.Error(t, err)
	})

	t.Run("renders specific page of multi-page PDF", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "multi.pdf")
		createTestPDFMultiPage(t, path, 3)

		png, err := RenderPage(path, 2)
		require.NoError(t, err)
		require.NotEmpty(t, png)
		assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, png[:4])
	})
}
