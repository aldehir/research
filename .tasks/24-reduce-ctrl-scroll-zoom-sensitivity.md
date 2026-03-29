# Task 24: Reduce ctrl+scroll zoom sensitivity

Lower the scaling factor for ctrl+scrollwheel zoom so jumps are less dramatic.

## Context

- `frontend/src/lib/pdf-utils.ts` defines `zoomByDelta(scale, deltaY)` which uses `deltaY * 0.002` as the multiplier — this is too aggressive
- The wheel handler is in `frontend/src/lib/PdfViewer.svelte` (`handleWheel` ~line 395), gated on `ctrlKey || metaKey`
- Button zoom uses `ZOOM_STEP = 0.25` (separate path, not affected)
- Scale is clamped to `[MIN_SCALE=0.25, MAX_SCALE=5.0]` and rounded to 2 decimals

## Checklist

- [x] Write test for `zoomByDelta` verifying smaller zoom increments per scroll tick
- [x] Reduce the `0.002` factor in `zoomByDelta` (try `0.001` or lower)
- [x] Manually verify smooth zoom feel in browser

## Notes

- The `deltaY` value varies by OS/browser/input device — trackpad pinch-zoom sends many small deltas while a mouse wheel sends larger discrete ones. The factor should feel smooth for both.
- Consider whether `deltaMode` (pixel vs line vs page) needs handling — most browsers normalize to pixels with `ctrlKey` but line-mode wheels could still send large values.
