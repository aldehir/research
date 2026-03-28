package pdf

import (
	"fmt"

	gopdf "github.com/ledongthuc/pdf"
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
	f, r, err := gopdf.Open(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	info := r.Trailer().Key("Info")
	return Metadata{
		Title:     info.Key("Title").Text(),
		Author:    info.Key("Author").Text(),
		Subject:   info.Key("Subject").Text(),
		CreatedAt: info.Key("CreationDate").Text(),
		PageCount: r.NumPage(),
	}, nil
}
