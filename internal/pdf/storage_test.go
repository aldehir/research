package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	t.Run("writes file and returns correct path and size", func(t *testing.T) {
		dir := t.TempDir()
		s := NewStorage(dir)

		content := "%PDF-1.4 test content"
		path, size, err := s.Save("abc-123", strings.NewReader(content))
		require.NoError(t, err)

		assert.Equal(t, filepath.Join(dir, "abc-123.pdf"), path)
		assert.Equal(t, int64(len(content)), size)

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("creates directory if missing", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nested", "dir")
		s := NewStorage(dir)

		_, _, err := s.Save("test-id", strings.NewReader("%PDF-1.4"))
		require.NoError(t, err)

		_, err = os.Stat(filepath.Join(dir, "test-id.pdf"))
		assert.NoError(t, err)
	})
}

func TestDelete(t *testing.T) {
	t.Run("removes existing file", func(t *testing.T) {
		dir := t.TempDir()
		s := NewStorage(dir)

		_, _, err := s.Save("del-id", strings.NewReader("%PDF-1.4"))
		require.NoError(t, err)

		err = s.Delete("del-id")
		require.NoError(t, err)

		_, err = os.Stat(s.Path("del-id"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("non-existent file is not an error", func(t *testing.T) {
		dir := t.TempDir()
		s := NewStorage(dir)

		err := s.Delete("does-not-exist")
		assert.NoError(t, err)
	})
}

func TestPath(t *testing.T) {
	s := NewStorage("/data/pdfs")
	assert.Equal(t, "/data/pdfs/my-id.pdf", s.Path("my-id"))
}
