package pdf

import (
	"os"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Metadata holds extracted PDF document metadata.
type Metadata struct {
	Title     string
	Author    string
	Subject   string
	CreatedAt string
	PageCount int
}

// ExtractMetadata reads PDF metadata from the file at the given path.
func ExtractMetadata(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, err
	}
	defer f.Close()

	info, err := pdfcpuapi.PDFInfo(f, path, nil, false, model.NewDefaultConfiguration())
	if err != nil {
		return Metadata{}, err
	}

	return Metadata{
		Title:     info.Title,
		Author:    info.Author,
		Subject:   info.Subject,
		CreatedAt: info.CreationDate,
		PageCount: info.PageCount,
	}, nil
}
