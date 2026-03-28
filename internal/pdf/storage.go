package pdf

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Storage manages PDF files on disk.
type Storage struct {
	Dir    string
	Logger *slog.Logger
}

// NewStorage creates a new Storage with the given base directory.
func NewStorage(dir string) *Storage {
	return &Storage{Dir: dir, Logger: slog.Default()}
}

func (s *Storage) logger() *slog.Logger {
	if s.Logger != nil {
		return s.Logger
	}
	return slog.Default()
}

// Save writes content to a file named {id}.pdf in the storage directory.
// Creates the directory if it doesn't exist.
func (s *Storage) Save(id string, content io.Reader) (string, int64, error) {
	log := s.logger()
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		log.Error("failed to create storage directory", "dir", s.Dir, "error", err)
		return "", 0, err
	}
	path := s.Path(id)
	f, err := os.Create(path)
	if err != nil {
		log.Error("failed to create file", "path", path, "error", err)
		return "", 0, err
	}
	defer f.Close()

	size, err := io.Copy(f, content)
	if err != nil {
		log.Error("failed to write file", "path", path, "error", err)
		return "", 0, err
	}
	log.Info("pdf saved", "id", id, "path", path, "size", size)
	return path, size, nil
}

// Delete removes the PDF file for the given ID.
// Returns nil if the file does not exist.
func (s *Storage) Delete(id string) error {
	log := s.logger()
	err := os.Remove(s.Path(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		log.Error("failed to delete pdf", "id", id, "error", err)
		return err
	}
	log.Info("pdf deleted", "id", id)
	return nil
}

// Path returns the full filesystem path for a paper's PDF.
func (s *Storage) Path(id string) string {
	return filepath.Join(s.Dir, id+".pdf")
}
