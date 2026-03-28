import type { PDFPageProxy } from 'pdfjs-dist';

export async function extractPageText(page: PDFPageProxy): Promise<string> {
	const content = await page.getTextContent();
	let text = '';
	for (const item of content.items) {
		if (!('str' in item)) continue;
		text += item.str;
		if (item.hasEOL) {
			text += '\n';
		}
	}
	return text;
}
