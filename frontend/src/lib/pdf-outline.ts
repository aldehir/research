import type { PDFDocumentProxy } from 'pdfjs-dist';

export interface TocEntry {
	title: string;
	pageNumber: number;
	children: TocEntry[];
}

interface OutlineItem {
	title: string;
	dest: string | unknown[] | null;
	items: OutlineItem[];
}

async function resolvePageNumber(
	doc: PDFDocumentProxy,
	dest: string | unknown[] | null
): Promise<number> {
	if (!dest) return 1;

	const resolved = typeof dest === 'string'
		? await doc.getDestination(dest)
		: dest;

	if (!resolved || resolved.length === 0) return 1;

	const ref = resolved[0];
	const pageIndex = await doc.getPageIndex(ref as any);
	return pageIndex + 1;
}

async function convertItems(
	doc: PDFDocumentProxy,
	items: OutlineItem[]
): Promise<TocEntry[]> {
	const entries: TocEntry[] = [];
	for (const item of items) {
		const pageNumber = await resolvePageNumber(doc, item.dest);
		const children = item.items?.length
			? await convertItems(doc, item.items)
			: [];
		entries.push({ title: item.title, pageNumber, children });
	}
	return entries;
}

/**
 * Find the TOC entry best matching the current page.
 * Walks the tree depth-first, returning the deepest entry whose
 * pageNumber <= currentPage.
 */
export function findActiveTocEntry(entries: TocEntry[], currentPage: number): TocEntry | null {
	if (entries.length === 0) return null;

	let best: TocEntry | null = null;

	function walk(items: TocEntry[]): void {
		for (const entry of items) {
			if (entry.pageNumber <= currentPage) {
				if (!best || entry.pageNumber >= best.pageNumber) {
					best = entry;
				}
				if (entry.children.length > 0) {
					walk(entry.children);
				}
			}
		}
	}

	walk(entries);

	// If no entry has pageNumber <= currentPage, return the first entry
	return best ?? entries[0];
}

export async function extractOutline(doc: PDFDocumentProxy): Promise<TocEntry[]> {
	const outline = await doc.getOutline();
	if (!outline || outline.length === 0) return [];
	return convertItems(doc, outline as OutlineItem[]);
}
