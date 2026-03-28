package pdf

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave_LogsOnSuccess(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	s := NewStorage(t.TempDir())
	s.Logger = logger

	_, _, err := s.Save("test-id", strings.NewReader("%PDF-1.4"))
	require.NoError(t, err)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "pdf saved")
	assert.Contains(t, logOutput, "test-id")
}

func TestDelete_LogsOnSuccess(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	s := NewStorage(t.TempDir())
	s.Logger = logger

	_, _, err := s.Save("del-id", strings.NewReader("%PDF-1.4"))
	require.NoError(t, err)

	buf.Reset()
	err = s.Delete("del-id")
	require.NoError(t, err)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "pdf deleted")
	assert.Contains(t, logOutput, "del-id")
}
