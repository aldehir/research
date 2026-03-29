// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock pdfjs-dist
vi.mock('pdfjs-dist', () => {
	class MockTextLayer {
		container: HTMLElement;
		constructor(opts: { container: HTMLElement }) {
			this.container = opts.container;
		}
		async render() {
			const span = document.createElement('span');
			span.textContent = 'hello';
			this.container.appendChild(span);
		}
	}
	return {
		TextLayer: MockTextLayer,
		GlobalWorkerOptions: { workerSrc: '' },
		getDocument: vi.fn()
	};
});

import { renderPage, getPageDimensions, PDF_TO_CSS_UNITS } from '$lib/pdf-render';

// Base unscaled PDF page dimensions (in PDF points)
const PAGE_WIDTH = 612;
const PAGE_HEIGHT = 792;

describe('renderPage', () => {
	let container: HTMLDivElement;

	const fakePage = {
		getViewport: ({ scale }: { scale: number }) => ({
			width: PAGE_WIDTH * scale,
			height: PAGE_HEIGHT * scale
		}),
		getTextContent: async () => ({ items: [], styles: {} }),
		render: () => ({ promise: Promise.resolve() })
	};

	beforeEach(() => {
		container = document.createElement('div');
		HTMLCanvasElement.prototype.getContext = vi.fn(() => ({
			scale: vi.fn(),
			fillRect: vi.fn(),
			fillStyle: ''
		})) as unknown as typeof HTMLCanvasElement.prototype.getContext;
	});

	it('creates viewport with scale * PDF_TO_CSS_UNITS', async () => {
		const getViewportSpy = vi.fn(({ scale }: { scale: number }) => ({
			width: PAGE_WIDTH * scale,
			height: PAGE_HEIGHT * scale
		}));
		const spyPage = { ...fakePage, getViewport: getViewportSpy };

		await renderPage(spyPage as any, container, 1.5);

		expect(getViewportSpy).toHaveBeenCalledWith({ scale: 1.5 * PDF_TO_CSS_UNITS });
	});

	it('stops rendering when abort signal fires after page.render', async () => {
		let resolveRender: () => void;
		const slowPage = {
			...fakePage,
			render: () => ({ promise: new Promise<void>((r) => { resolveRender = r; }) })
		};

		const ac = new AbortController();
		const promise = renderPage(slowPage as any, container, 1.0, ac.signal);

		// Canvas is appended synchronously before the await
		expect(container.querySelector('canvas')).not.toBeNull();

		// Abort while page.render is still pending
		ac.abort();
		resolveRender!();
		await promise;

		// textLayer should NOT be added since signal was aborted
		expect(container.querySelector('.textLayer')).toBeNull();
	});
});

describe('getPageDimensions', () => {
	const fakePage = {
		getViewport: ({ scale }: { scale: number }) => ({
			width: PAGE_WIDTH * scale,
			height: PAGE_HEIGHT * scale
		}),
		getTextContent: async () => ({ items: [], styles: {} }),
		render: () => ({ promise: Promise.resolve() })
	};

	it('returns dimensions including PDF_TO_CSS_UNITS at scale 1.0', () => {
		const dims = getPageDimensions(fakePage as any, 1.0);
		expect(dims.width).toBeCloseTo(PAGE_WIDTH * PDF_TO_CSS_UNITS, 4);
		expect(dims.height).toBeCloseTo(PAGE_HEIGHT * PDF_TO_CSS_UNITS, 4);
	});

	it('scales dimensions with scale factor and PDF_TO_CSS_UNITS', () => {
		const dims = getPageDimensions(fakePage as any, 2.0);
		expect(dims.width).toBeCloseTo(PAGE_WIDTH * 2.0 * PDF_TO_CSS_UNITS, 4);
		expect(dims.height).toBeCloseTo(PAGE_HEIGHT * 2.0 * PDF_TO_CSS_UNITS, 4);
	});
});
