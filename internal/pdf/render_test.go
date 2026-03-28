package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPage(t *testing.T) {
	t.Run("renders a page to PNG bytes", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.pdf")
		createTestPDFWithText(t, path, "Hello World")

		pngBytes, err := RenderPage(path, 1)
		require.NoError(t, err)
		require.NotEmpty(t, pngBytes)
		// PNG files start with the 8-byte magic header
		assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, pngBytes[:8])
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

		pngBytes, err := RenderPage(path, 2)
		require.NoError(t, err)
		require.NotEmpty(t, pngBytes)
		assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, pngBytes[:4])
	})

	t.Run("cropped output is smaller than full page", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.pdf")
		// Text in top-left corner — most of the page is whitespace
		createTestPDFWithText(t, path, "Hello")

		pngBytes, err := RenderPage(path, 1)
		require.NoError(t, err)

		img, err := png.Decode(bytes.NewReader(pngBytes))
		require.NoError(t, err)

		bounds := img.Bounds()
		// Letter at 150 DPI = 1275x1650. Cropped should be much smaller.
		assert.Less(t, bounds.Dx(), 600, "cropped width should be much less than full page width")
		assert.Less(t, bounds.Dy(), 400, "cropped height should be much less than full page height")
	})
}

func TestCropWhitespace(t *testing.T) {
	t.Run("crops to content bounding box with padding", func(t *testing.T) {
		// 100x100 white image with a 10x10 black block at (40,40)-(50,50)
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				img.Set(x, y, color.White)
			}
		}
		for y := 40; y < 50; y++ {
			for x := 40; x < 50; x++ {
				img.Set(x, y, color.Black)
			}
		}

		cropped := cropWhitespace(img, 5)
		bounds := cropped.Bounds()

		// Content is (40,40)-(50,50), padding 5 → expect (35,35)-(55,55) = 20x20
		assert.Equal(t, 20, bounds.Dx())
		assert.Equal(t, 20, bounds.Dy())
	})

	t.Run("padding is clamped to image bounds", func(t *testing.T) {
		// 20x20 white image with a 2x2 block at (1,1)-(3,3)
		img := image.NewRGBA(image.Rect(0, 0, 20, 20))
		for y := 0; y < 20; y++ {
			for x := 0; x < 20; x++ {
				img.Set(x, y, color.White)
			}
		}
		for y := 1; y < 3; y++ {
			for x := 1; x < 3; x++ {
				img.Set(x, y, color.Black)
			}
		}

		cropped := cropWhitespace(img, 10)
		bounds := cropped.Bounds()

		// Padding 10 would exceed (0,0) on the left/top → clamped to image bounds
		assert.LessOrEqual(t, bounds.Min.X, 0)
		assert.LessOrEqual(t, bounds.Min.Y, 0)
		assert.GreaterOrEqual(t, bounds.Max.X, 3)
		assert.GreaterOrEqual(t, bounds.Max.Y, 3)
	})

	t.Run("all-white image returns original", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 50, 50))
		for y := 0; y < 50; y++ {
			for x := 0; x < 50; x++ {
				img.Set(x, y, color.White)
			}
		}

		cropped := cropWhitespace(img, 5)
		bounds := cropped.Bounds()

		assert.Equal(t, 50, bounds.Dx())
		assert.Equal(t, 50, bounds.Dy())
	})
}
