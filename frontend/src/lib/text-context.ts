export function extractPageText(pageElement: HTMLDivElement): string {
	return pageElement.textContent ?? '';
}

export function extractSurroundingContext(
	selectedText: string,
	pageText: string,
	contextWindow: number = 500
): string {
	const index = pageText.indexOf(selectedText);
	if (index === -1) {
		return pageText;
	}

	const start = Math.max(0, index - contextWindow);
	const end = Math.min(pageText.length, index + selectedText.length + contextWindow);
	return pageText.slice(start, end);
}
