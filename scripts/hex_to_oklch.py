#!/usr/bin/env python3
"""Convert hex and rgba CSS colors to OKLCH format.

Usage:
    python3 hex_to_oklch.py              # prints conversion table for theme.css colors
    python3 hex_to_oklch.py '#2563eb'    # converts a single hex color
"""
import math
import re
import sys


def _srgb_to_linear(c: float) -> float:
    """Convert sRGB component (0-1) to linear RGB."""
    if c <= 0.04045:
        return c / 12.92
    return ((c + 0.055) / 1.055) ** 2.4


def _linear_to_xyz(r: float, g: float, b: float) -> tuple[float, float, float]:
    """Linear RGB to CIE XYZ (D65)."""
    x = 0.4124564 * r + 0.3575761 * g + 0.1804375 * b
    y = 0.2126729 * r + 0.7151522 * g + 0.0721750 * b
    z = 0.0193339 * r + 0.1191920 * g + 0.9503041 * b
    return x, y, z


def _xyz_to_oklab(x: float, y: float, z: float) -> tuple[float, float, float]:
    """CIE XYZ to OKLab."""
    l_ = 0.8189330101 * x + 0.3618667424 * y - 0.1288597137 * z
    m_ = 0.0329845436 * x + 0.9293118715 * y + 0.0361456387 * z
    s_ = 0.0482003018 * x + 0.2643662691 * y + 0.6338517070 * z

    l_c = math.copysign(abs(l_) ** (1 / 3), l_) if l_ != 0 else 0
    m_c = math.copysign(abs(m_) ** (1 / 3), m_) if m_ != 0 else 0
    s_c = math.copysign(abs(s_) ** (1 / 3), s_) if s_ != 0 else 0

    L = 0.2104542553 * l_c + 0.7936177850 * m_c - 0.0040720468 * s_c
    a = 1.9779984951 * l_c - 2.4285922050 * m_c + 0.4505937099 * s_c
    b = 0.0259040371 * l_c + 0.7827717662 * m_c - 0.8086757660 * s_c
    return L, a, b


def _oklab_to_oklch(L: float, a: float, b: float) -> tuple[float, float, float]:
    """OKLab to OKLCH."""
    C = math.sqrt(a * a + b * b)
    H = math.degrees(math.atan2(b, a)) % 360
    return L, C, H


def _fmt(v: float, places: int = 4) -> str:
    """Format a float, stripping trailing zeros."""
    return f"{v:.{places}f}".rstrip("0").rstrip(".")


def hex_to_oklch(hex_color: str) -> str:
    """Convert a hex color string to oklch() CSS format."""
    h = hex_color.lstrip("#")
    if len(h) == 3:
        h = "".join(c * 2 for c in h)

    r_i, g_i, b_i = int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16)

    r = _srgb_to_linear(r_i / 255)
    g = _srgb_to_linear(g_i / 255)
    b = _srgb_to_linear(b_i / 255)

    x, y, z = _linear_to_xyz(r, g, b)
    L, a, ob = _xyz_to_oklab(x, y, z)
    L, C, H = _oklab_to_oklch(L, a, ob)

    # Round for CSS output
    L = round(L, 4)
    C = round(C, 4)
    H = round(H, 2)

    # Clean up near-zero chroma (achromatic colors)
    if C < 0.0005:
        C = 0
        H = 0

    return f"oklch({_fmt(L)} {_fmt(C)} {_fmt(H)})"


def rgba_to_oklch(rgba_str: str) -> str:
    """Convert an rgba() CSS string to oklch() with alpha."""
    m = re.match(
        r"rgba\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*,\s*([0-9.]+)\s*\)",
        rgba_str,
    )
    if not m:
        raise ValueError(f"Cannot parse rgba: {rgba_str}")

    r_i, g_i, b_i = int(m.group(1)), int(m.group(2)), int(m.group(3))
    alpha = m.group(4)

    hex_color = f"#{r_i:02x}{g_i:02x}{b_i:02x}"
    base = hex_to_oklch(hex_color)
    # Insert alpha: oklch(L C H) -> oklch(L C H / alpha)
    return base.replace(")", f" / {alpha})")


def convert_css_file(path: str) -> str:
    """Read a CSS file and replace all hex/rgba colors with oklch equivalents."""
    with open(path) as f:
        content = f.read()

    # Replace hex colors (6-digit and 3-digit)
    def replace_hex(m: re.Match) -> str:
        return hex_to_oklch(m.group(0))

    content = re.sub(r"#[0-9a-fA-F]{6}\b", replace_hex, content)
    content = re.sub(r"#[0-9a-fA-F]{3}\b", replace_hex, content)

    # Replace rgba() values
    def replace_rgba(m: re.Match) -> str:
        return rgba_to_oklch(m.group(0))

    content = re.sub(
        r"rgba\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*,\s*[0-9.]+\s*\)",
        replace_rgba,
        content,
    )

    return content


if __name__ == "__main__":
    if len(sys.argv) > 1:
        arg = sys.argv[1]
        if arg.startswith("#"):
            print(f"{arg} -> {hex_to_oklch(arg)}")
        elif arg.startswith("rgba"):
            print(f"{arg} -> {rgba_to_oklch(arg)}")
        else:
            # Treat as file path
            print(convert_css_file(arg))
    else:
        print("Usage: python3 hex_to_oklch.py '#2563eb'")
        print("       python3 hex_to_oklch.py 'rgba(0, 0, 0, 0.3)'")
        print("       python3 hex_to_oklch.py path/to/file.css")
