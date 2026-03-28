import { TextLayer } from 'pdfjs-dist';
import type { PDFPageProxy } from 'pdfjs-dist';

export async function renderPage(
	page: PDFPageProxy,
	container: HTMLDivElement,
	currentScale: number
): Promise<void> {
	const viewport = page.getViewport({ scale: currentScale });

	container.innerHTML = '';
	container.style.width = `${viewport.width}px`;
	container.style.height = `${viewport.height}px`;
	container.style.position = 'relative';

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

	await page.render({ canvasContext: ctx, viewport }).promise;

	const textContent = await page.getTextContent();
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
