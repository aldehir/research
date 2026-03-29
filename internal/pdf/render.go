package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os/exec"
	"strconv"

	xdraw "golang.org/x/image/draw"
)

// RenderPage renders a single PDF page to a PNG image using pdftoppm.
// pageNum is 1-based. Returns the raw PNG bytes, cropped to content
// with a small margin to minimize wasted whitespace.
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

	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		return nil, fmt.Errorf("decode png: %w", err)
	}

	cropped := cropWhitespace(img, 20)
	final := constrainSize(cropped, maxImageDimension)

	var buf bytes.Buffer
	if err := png.Encode(&buf, final); err != nil {
		return nil, fmt.Errorf("encode cropped png: %w", err)
	}
	return buf.Bytes(), nil
}

// RenderRegion renders a rectangular region of a PDF page to a PNG image.
// Coordinates (x, y, w, h) are in PDF points (1/72 inch), origin at top-left.
// Renders at 150 DPI. No whitespace cropping since the user chose the bounds.
func RenderRegion(path string, pageNum, x, y, w, h int) ([]byte, error) {
	count, err := PageCount(path)
	if err != nil {
		return nil, fmt.Errorf("render region: %w", err)
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
		"-x", strconv.Itoa(x),
		"-y", strconv.Itoa(y),
		"-W", strconv.Itoa(w),
		"-H", strconv.Itoa(h),
		"-singlefile",
		path,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm region: %w", err)
	}

	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		return nil, fmt.Errorf("decode png: %w", err)
	}

	final := constrainSize(img, maxImageDimension)

	var buf bytes.Buffer
	if err := png.Encode(&buf, final); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// maxImageDimension is the maximum width or height for rendered images.
// Matches Anthropic's recommended tile size to avoid unnecessary scaling.
const maxImageDimension = 1568

// constrainSize scales the image down proportionally if either dimension
// exceeds maxPx. Returns the original image if already within bounds.
func constrainSize(img image.Image, maxPx int) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	if w <= maxPx && h <= maxPx {
		return img
	}

	scale := float64(maxPx) / float64(max(w, h))
	newW := int(float64(w) * scale)
	newH := int(float64(h) * scale)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}

// cropWhitespace finds the bounding box of non-white pixels and returns
// a sub-image cropped to that region plus the given padding on each side.
// If the image is entirely white, returns it unchanged.
func cropWhitespace(img image.Image, padding int) image.Image {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Check if pixel is not white (using threshold to handle
			// near-white antialiased edges)
			if r < 0xF000 || g < 0xF000 || b < 0xF000 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	// All white — return original
	if minX > maxX || minY > maxY {
		return img
	}

	// Expand by padding, clamped to image bounds
	cropRect := image.Rect(
		max(bounds.Min.X, minX-padding),
		max(bounds.Min.Y, minY-padding),
		min(bounds.Max.X, maxX+1+padding),
		min(bounds.Max.Y, maxY+1+padding),
	)

	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if si, ok := img.(subImager); ok {
		return si.SubImage(cropRect)
	}
	return img
}
