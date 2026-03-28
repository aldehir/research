// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * Tests for PDF page rendering structure.
 * Verifies that renderPage produces the correct DOM hierarchy:
 *   container > canvas + div.textLayer
 * and that dimensions are set from the viewport.
 */

// Mock pdfjs-dist TextLayer
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

// Import the render helper (will be extracted from PdfViewer)
import { renderPage } from '$lib/pdf-render';

describe('renderPage', () => {
	let container: HTMLDivElement;
	const fakeViewport = {
		width: 612,
		height: 792
	};

	const fakePage = {
		getViewport: () => fakeViewport,
		getTextContent: async () => ({ items: [], styles: {} }),
		render: () => ({ promise: Promise.resolve() })
	};

	beforeEach(() => {
		container = document.createElement('div');
		// jsdom doesn't implement getContext, mock it
		HTMLCanvasElement.prototype.getContext = vi.fn(() => ({
			scale: vi.fn()
		})) as unknown as typeof HTMLCanvasElement.prototype.getContext;
	});

	it('creates canvas and textLayer div inside container', async () => {
		await renderPage(fakePage as any, container, 1.0);

		const canvas = container.querySelector('canvas');
		expect(canvas).not.toBeNull();

		const textLayer = container.querySelector('.textLayer');
		expect(textLayer).not.toBeNull();
		expect(textLayer!.tagName).toBe('DIV');
	});

	it('sets container dimensions from viewport', async () => {
		await renderPage(fakePage as any, container, 1.0);

		expect(container.style.width).toBe('612px');
		expect(container.style.height).toBe('792px');
	});

	it('sets canvas CSS dimensions to match viewport', async () => {
		await renderPage(fakePage as any, container, 1.0);

		const canvas = container.querySelector('canvas')!;
		expect(canvas.style.width).toBe('612px');
		expect(canvas.style.height).toBe('792px');
	});

	it('renders text content into textLayer', async () => {
		await renderPage(fakePage as any, container, 1.0);

		const textLayer = container.querySelector('.textLayer')!;
		const spans = textLayer.querySelectorAll('span');
		expect(spans.length).toBeGreaterThan(0);
	});

	it('does not apply conflicting inline styles to textLayer', async () => {
		await renderPage(fakePage as any, container, 1.0);

		const textLayer = container.querySelector('.textLayer') as HTMLDivElement;
		// textLayer positioning should come from pdf_viewer.css, not inline styles
		expect(textLayer.style.position).toBe('');
		expect(textLayer.style.lineHeight).toBe('');
		expect(textLayer.style.overflow).toBe('');
	});
});
