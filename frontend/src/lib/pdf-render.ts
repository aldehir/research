import { TextLayer, AnnotationLayer } from 'pdfjs-dist';
import type { PDFPageProxy } from 'pdfjs-dist';

export const PDF_TO_CSS_UNITS = 96 / 72;

export function getPageDimensions(
	page: PDFPageProxy,
	scale: number
): { width: number; height: number } {
	const viewport = page.getViewport({ scale: scale * PDF_TO_CSS_UNITS });
	return { width: viewport.width, height: viewport.height };
}

export function clearPage(container: HTMLDivElement): void {
	container.innerHTML = '';
}

export async function renderAnnotations(
	page: PDFPageProxy,
	container: HTMLDivElement,
	currentScale: number,
	linkService?: unknown
): Promise<void> {
	const annotations = await page.getAnnotations();
	if (annotations.length === 0) return;

	const viewport = page.getViewport({ scale: currentScale * PDF_TO_CSS_UNITS });
	const annotDiv = document.createElement('div');
	annotDiv.className = 'annotationLayer';
	container.appendChild(annotDiv);

	const layer = new AnnotationLayer({
		div: annotDiv,
		page,
		viewport,
		accessibilityManager: null,
		annotationCanvasMap: null,
		annotationEditorUIManager: null,
		structTreeLayer: null,
		commentManager: null,
		linkService: linkService ?? null,
		annotationStorage: null
	});

	await layer.render({
		viewport,
		div: annotDiv,
		annotations,
		page,
		linkService: linkService ?? null,
		renderForms: false
	} as any);
}

export async function renderPage(
	page: PDFPageProxy,
	container: HTMLDivElement,
	currentScale: number,
	signal?: AbortSignal
): Promise<void> {
	const viewport = page.getViewport({ scale: currentScale * PDF_TO_CSS_UNITS });

	container.innerHTML = '';
	container.style.width = `${viewport.width}px`;
	container.style.height = `${viewport.height}px`;
	container.style.position = 'relative';

	// pdf.js TextLayer CSS uses --total-scale-factor for font sizing and
	// container dimensions. Normally set by PDFViewer on .pdfViewer .page,
	// but we use TextLayer standalone so must set it ourselves.
	container.style.setProperty(
		'--total-scale-factor',
		`${currentScale * PDF_TO_CSS_UNITS}`
	);

	const canvas = document.createElement('canvas');
	const dpr = typeof window !== 'undefined' ? (window.devicePixelRatio || 1) : 1;
	canvas.width = Math.floor(viewport.width * dpr);
	canvas.height = Math.floor(viewport.height * dpr);
	canvas.style.width = `${viewport.width}px`;
	canvas.style.height = `${viewport.height}px`;

	const ctx = canvas.getContext('2d');
	if (!ctx) return;

	ctx.scale(dpr, dpr);
	container.appendChild(canvas);

	await page.render({ canvasContext: ctx, canvas, viewport }).promise;
	if (signal?.aborted) return;

	const textContent = await page.getTextContent();
	if (signal?.aborted) return;

	const textDiv = document.createElement('div');
	textDiv.className = 'textLayer';
	container.appendChild(textDiv);

	const textLayer = new TextLayer({
		textContentSource: textContent,
		container: textDiv,
		viewport
	});
	await textLayer.render();
}
