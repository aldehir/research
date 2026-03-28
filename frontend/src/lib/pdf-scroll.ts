/**
 * Scroll anchoring utilities for virtual PDF rendering.
 * Pure functions for computing and restoring scroll position during zoom/rerender.
 */

export interface ScrollAnchor {
	/** 1-based page number at the top of the viewport */
	pageNum: number;
	/** Fraction of the anchor page that is above the viewport top (0..1) */
	offset: number;
}

/**
 * Compute a scroll anchor from the current scroll state.
 * Finds the page closest to the viewport top and records how far into
 * that page the viewport top sits.
 */
export function computeScrollAnchor(
	scrollTop: number,
	pageOffsets: { pageNum: number; top: number; height: number }[]
): ScrollAnchor {
	if (pageOffsets.length === 0) {
		return { pageNum: 1, offset: 0 };
	}

	// Find the page that contains the viewport top
	for (let i = pageOffsets.length - 1; i >= 0; i--) {
		const p = pageOffsets[i];
		if (p.top <= scrollTop) {
			const offset = p.height > 0 ? (scrollTop - p.top) / p.height : 0;
			return { pageNum: p.pageNum, offset: Math.min(offset, 1) };
		}
	}

	return { pageNum: pageOffsets[0].pageNum, offset: 0 };
}

/**
 * Compute the new scrollTop to restore a scroll anchor after pages have been resized.
 */
export function restoreScrollTop(
	anchor: ScrollAnchor,
	newPageOffsets: { pageNum: number; top: number; height: number }[]
): number {
	const page = newPageOffsets.find((p) => p.pageNum === anchor.pageNum);
	if (!page) return 0;
	return page.top + page.height * anchor.offset;
}
