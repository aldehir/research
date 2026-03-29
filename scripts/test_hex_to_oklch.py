"""Tests for hex/rgba to oklch conversion."""
import unittest
from hex_to_oklch import hex_to_oklch, rgba_to_oklch


class TestHexToOklch(unittest.TestCase):
    def test_white(self):
        self.assertEqual(hex_to_oklch("#ffffff"), "oklch(1 0 0)")

    def test_black(self):
        self.assertEqual(hex_to_oklch("#000000"), "oklch(0 0 0)")

    def test_pure_red(self):
        result = hex_to_oklch("#ff0000")
        # Should have high lightness, high chroma, hue ~29
        self.assertTrue(result.startswith("oklch("))
        parts = result.replace("oklch(", "").replace(")", "").split()
        L, C, H = float(parts[0]), float(parts[1]), float(parts[2])
        self.assertAlmostEqual(L, 0.6279, places=2)
        self.assertAlmostEqual(H, 29.23, places=0)
        self.assertGreater(C, 0.2)

    def test_primary_blue(self):
        result = hex_to_oklch("#2563eb")
        self.assertTrue(result.startswith("oklch("))
        parts = result.replace("oklch(", "").replace(")", "").split()
        L, C, H = float(parts[0]), float(parts[1]), float(parts[2])
        # Blue hue should be roughly 260-270
        self.assertGreater(H, 250)
        self.assertLess(H, 280)

    def test_shorthand_hex(self):
        # 3-char hex should work too
        result = hex_to_oklch("#fff")
        self.assertEqual(result, "oklch(1 0 0)")

    def test_gray(self):
        result = hex_to_oklch("#808080")
        parts = result.replace("oklch(", "").replace(")", "").split()
        L, C = float(parts[0]), float(parts[1])
        # Gray should have near-zero chroma
        self.assertLess(C, 0.01)
        self.assertAlmostEqual(L, 0.6, places=1)


class TestRgbaToOklch(unittest.TestCase):
    def test_black_with_alpha(self):
        self.assertEqual(rgba_to_oklch("rgba(0, 0, 0, 0.3)"), "oklch(0 0 0 / 0.3)")

    def test_white_with_alpha(self):
        self.assertEqual(rgba_to_oklch("rgba(255, 255, 255, 0.08)"), "oklch(1 0 0 / 0.08)")

    def test_color_with_alpha(self):
        result = rgba_to_oklch("rgba(59, 130, 246, 0.15)")
        self.assertTrue(result.startswith("oklch("))
        self.assertTrue("/ 0.15" in result)

    def test_white_with_0_1_alpha(self):
        self.assertEqual(rgba_to_oklch("rgba(255, 255, 255, 0.1)"), "oklch(1 0 0 / 0.1)")


if __name__ == "__main__":
    unittest.main()
