// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * Tests for PDF page rendering structure.
 * Verifies that renderPage produces the correct DOM hierarchy:
 *   container > canvas + div.textLayer
 * and that dimensions are set from the viewport.
 */

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
	class MockAnnotationLayer {
		div: HTMLElement;
		constructor(opts: { div: HTMLElement }) {
			this.div = opts.div;
		}
		async render(params: { annotations: any[] }) {
			for (const ann of params.annotations) {
				const section = document.createElement('section');
				section.dataset.annotationType = String(ann.annotationType);
				this.div.appendChild(section);
			}
		}
	}
	return {
		TextLayer: MockTextLayer,
		AnnotationLayer: MockAnnotationLayer,
		GlobalWorkerOptions: { workerSrc: '' },
		getDocument: vi.fn()
	};
});

// Import the render helper (will be extracted from PdfViewer)
import { renderPage, renderAnnotations, clearPage, getPageDimensions } from '$lib/pdf-render';

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

	it('sets --total-scale-factor CSS variable on container', async () => {
		await renderPage(fakePage as any, container, 1.5);

		const value = container.style.getPropertyValue('--total-scale-factor');
		expect(value).not.toBe('');
		// scale * PDF_TO_CSS_UNITS (96/72 = 1.333...)
		expect(parseFloat(value)).toBeCloseTo(1.5 * (96 / 72), 4);
	});

	it('sets --total-scale-factor at default scale', async () => {
		await renderPage(fakePage as any, container, 1.0);

		const value = container.style.getPropertyValue('--total-scale-factor');
		expect(parseFloat(value)).toBeCloseTo(96 / 72, 4);
	});
});

describe('clearPage', () => {
	let container: HTMLDivElement;

	beforeEach(() => {
		container = document.createElement('div');
		container.style.width = '612px';
		container.style.height = '792px';
		container.style.position = 'relative';

		const canvas = document.createElement('canvas');
		const textDiv = document.createElement('div');
		textDiv.className = 'textLayer';
		container.appendChild(canvas);
		container.appendChild(textDiv);
	});

	it('removes all child elements', () => {
		clearPage(container);
		expect(container.children.length).toBe(0);
	});

	it('preserves container dimensions', () => {
		clearPage(container);
		expect(container.style.width).toBe('612px');
		expect(container.style.height).toBe('792px');
	});

	it('preserves container position', () => {
		clearPage(container);
		expect(container.style.position).toBe('relative');
	});
});

describe('getPageDimensions', () => {
	const fakeViewport = { width: 612, height: 792 };
	const fakePage = {
		getViewport: ({ scale }: { scale: number }) => ({
			width: 612 * scale,
			height: 792 * scale
		}),
		getTextContent: async () => ({ items: [], styles: {} }),
		render: () => ({ promise: Promise.resolve() })
	};

	it('returns width and height at given scale', () => {
		const dims = getPageDimensions(fakePage as any, 1.0);
		expect(dims.width).toBe(612);
		expect(dims.height).toBe(792);
	});

	it('scales dimensions with scale factor', () => {
		const dims = getPageDimensions(fakePage as any, 2.0);
		expect(dims.width).toBe(1224);
		expect(dims.height).toBe(1584);
	});
});

describe('renderAnnotations', () => {
	let container: HTMLDivElement;
	const fakeViewport = { width: 612, height: 792 };
	const fakePage = {
		getViewport: () => fakeViewport,
		getAnnotations: async () => [
			{ annotationType: 2, url: 'https://example.com' },
			{ annotationType: 2, dest: [1, { name: 'Fit' }] }
		]
	};

	beforeEach(() => {
		container = document.createElement('div');
		container.style.position = 'relative';
	});

	it('creates annotationLayer div inside container', async () => {
		await renderAnnotations(fakePage as any, container, 1.0);

		const annotationLayer = container.querySelector('.annotationLayer');
		expect(annotationLayer).not.toBeNull();
		expect(annotationLayer!.tagName).toBe('DIV');
	});

	it('renders annotation elements', async () => {
		await renderAnnotations(fakePage as any, container, 1.0);

		const annotationLayer = container.querySelector('.annotationLayer')!;
		const sections = annotationLayer.querySelectorAll('section');
		expect(sections.length).toBe(2);
	});

	it('does not create layer when page has no annotations', async () => {
		const emptyPage = {
			...fakePage,
			getAnnotations: async () => []
		};
		await renderAnnotations(emptyPage as any, container, 1.0);

		const annotationLayer = container.querySelector('.annotationLayer');
		expect(annotationLayer).toBeNull();
	});
});
