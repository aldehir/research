package pdf

import (
	"fmt"
	"os/exec"
	"strconv"
)

// RenderPage renders a single PDF page to a PNG image using pdftoppm.
// pageNum is 1-based. Returns the raw PNG bytes.
// Renders at 150 DPI to balance quality vs size.
func RenderPage(path string, pageNum int) ([]byte, error) {
	count, err := PageCount(path)
	if err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}
	if pageNum < 1 || pageNum > count {
		return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, count)
	}

	pageStr := strconv.Itoa(pageNum)
	out, err := exec.Command(
		"pdftoppm",
		"-png",
		"-r", "150",
		"-f", pageStr,
		"-l", pageStr,
		"-singlefile",
		path,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm: %w", err)
	}
	return out, nil
}
