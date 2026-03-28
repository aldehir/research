package pdf

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Storage manages PDF files on disk.
type Storage struct {
	Dir string
}

// NewStorage creates a new Storage with the given base directory.
func NewStorage(dir string) *Storage {
	return &Storage{Dir: dir}
}

// Save writes content to a file named {id}.pdf in the storage directory.
// Creates the directory if it doesn't exist.
func (s *Storage) Save(id string, content io.Reader) (string, int64, error) {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return "", 0, err
	}
	path := s.Path(id)
	f, err := os.Create(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	size, err := io.Copy(f, content)
	if err != nil {
		return "", 0, err
	}
	return path, size, nil
}

// Delete removes the PDF file for the given ID.
// Returns nil if the file does not exist.
func (s *Storage) Delete(id string) error {
	err := os.Remove(s.Path(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Path returns the full filesystem path for a paper's PDF.
func (s *Storage) Path(id string) string {
	return filepath.Join(s.Dir, id+".pdf")
}
