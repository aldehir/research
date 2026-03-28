import type { PDFPageProxy } from 'pdfjs-dist';
import { extractPageText } from '$lib/pdf-text';

let selectedText = $state('');
let currentPage = $state(1);
let pageProxies = $state<PDFPageProxy[]>([]);

export function setSelectedText(text: string): void {
	selectedText = text;
}

export function getSelectedText(): string {
	return selectedText;
}

export function clearSelectedText(): void {
	selectedText = '';
}

export function setCurrentPage(page: number): void {
	currentPage = page;
}

export function getCurrentPage(): number {
	return currentPage;
}

export function setPages(pages: PDFPageProxy[]): void {
	pageProxies = pages;
}

export async function getSurroundingText(): Promise<string> {
	if (pageProxies.length === 0) return '';

	const start = Math.max(0, currentPage - 2); // prev page (0-indexed)
	const end = Math.min(pageProxies.length, currentPage + 1); // next page + 1

	const texts: string[] = [];
	for (let i = start; i < end; i++) {
		texts.push(await extractPageText(pageProxies[i]));
	}
	return texts.join('\n');
}
