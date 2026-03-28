// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest';
import { computeScrollAnchor, restoreScrollTop } from '$lib/pdf-scroll';

// Mock pdfjs-dist to avoid DOMMatrix error in jsdom
vi.mock('pdfjs-dist', () => ({
	TextLayer: class { async render() {} },
	AnnotationLayer: class { async render() {} },
	GlobalWorkerOptions: { workerSrc: '' },
	getDocument: vi.fn()
}));

import { getPageDimensions, PDF_TO_CSS_UNITS } from '$lib/pdf-render';

// Simulated page dimensions (standard US Letter in PDF points)
const PAGE_WIDTH = 612;
const PAGE_HEIGHT = 792;
const GAP = 8; // gap between pages in the container

function makePageOffsets(
	count: number,
	pageHeight: number,
	gap: number = GAP
): { pageNum: number; top: number; height: number }[] {
	const offsets = [];
	let top = 0;
	for (let i = 1; i <= count; i++) {
		offsets.push({ pageNum: i, top, height: pageHeight });
		top += pageHeight + gap;
	}
	return offsets;
}

describe('computeScrollAnchor', () => {
	it('returns page 1 offset 0 when scrolled to top', () => {
		const offsets = makePageOffsets(5, 800);
		const anchor = computeScrollAnchor(0, offsets);
		expect(anchor.pageNum).toBe(1);
		expect(anchor.offset).toBe(0);
	});

	it('returns correct page when scrolled partway through', () => {
		const offsets = makePageOffsets(5, 800);
		// Scroll to middle of page 2: page2.top = 808, middle = 808 + 400 = 1208
		const anchor = computeScrollAnchor(1208, offsets);
		expect(anchor.pageNum).toBe(2);
		expect(anchor.offset).toBeCloseTo(0.5, 2);
	});

	it('returns last page when scrolled past all pages', () => {
		const offsets = makePageOffsets(3, 800);
		const anchor = computeScrollAnchor(99999, offsets);
		expect(anchor.pageNum).toBe(3);
		expect(anchor.offset).toBe(1);
	});

	it('returns page 1 offset 0 for empty offsets', () => {
		const anchor = computeScrollAnchor(500, []);
		expect(anchor.pageNum).toBe(1);
		expect(anchor.offset).toBe(0);
	});

	it('handles scroll position exactly at page boundary', () => {
		const offsets = makePageOffsets(5, 800);
		// Page 3 starts at (800+8)*2 = 1616
		const anchor = computeScrollAnchor(1616, offsets);
		expect(anchor.pageNum).toBe(3);
		expect(anchor.offset).toBe(0);
	});
});

describe('restoreScrollTop', () => {
	it('restores to top of page when offset is 0', () => {
		const offsets = makePageOffsets(5, 1000);
		const result = restoreScrollTop({ pageNum: 3, offset: 0 }, offsets);
		expect(result).toBe(offsets[2].top);
	});

	it('restores proportional position within page', () => {
		const offsets = makePageOffsets(5, 1000);
		const result = restoreScrollTop({ pageNum: 2, offset: 0.5 }, offsets);
		expect(result).toBe(offsets[1].top + 500);
	});

	it('returns 0 when anchor page is not found', () => {
		const offsets = makePageOffsets(3, 800);
		const result = restoreScrollTop({ pageNum: 10, offset: 0.5 }, offsets);
		expect(result).toBe(0);
	});
});

describe('scroll anchor round-trip', () => {
	it('preserves scroll position when page sizes do not change', () => {
		const offsets = makePageOffsets(10, 800);
		const scrollTop = 2500;
		const anchor = computeScrollAnchor(scrollTop, offsets);
		const restored = restoreScrollTop(anchor, offsets);
		expect(restored).toBeCloseTo(scrollTop, 0);
	});

	it('adjusts scroll position proportionally when pages resize (zoom)', () => {
		const oldHeight = 800;
		const newHeight = 1200; // 1.5x zoom
		const oldOffsets = makePageOffsets(10, oldHeight);
		const newOffsets = makePageOffsets(10, newHeight);

		// Scroll to middle of page 3
		const page3Old = oldOffsets[2];
		const scrollTop = page3Old.top + oldHeight * 0.5;

		const anchor = computeScrollAnchor(scrollTop, oldOffsets);
		expect(anchor.pageNum).toBe(3);
		expect(anchor.offset).toBeCloseTo(0.5, 2);

		const restored = restoreScrollTop(anchor, newOffsets);
		const page3New = newOffsets[2];
		expect(restored).toBeCloseTo(page3New.top + newHeight * 0.5, 0);
	});
});

describe('placeholder vs rendered dimensions', () => {
	// Both getPageDimensions and renderPage use the same viewport calculation:
	//   page.getViewport({ scale: currentScale * PDF_TO_CSS_UNITS })
	// This test verifies they produce identical values.

	const fakePage = {
		getViewport: ({ scale }: { scale: number }) => ({
			width: PAGE_WIDTH * scale,
			height: PAGE_HEIGHT * scale
		})
	};

	const scales = [0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 2.0, 3.0, 5.0];

	for (const s of scales) {
		it(`getPageDimensions matches viewport at scale ${s}`, () => {
			const dims = getPageDimensions(fakePage as any, s);
			const viewport = fakePage.getViewport({ scale: s * PDF_TO_CSS_UNITS });
			expect(dims.width).toBe(viewport.width);
			expect(dims.height).toBe(viewport.height);
		});
	}
});
