// Shared navigation state between chat and PDF viewer.
// Chat sets a target page, PdfViewer observes and scrolls to it.

let navigateTarget = $state<number | null>(null);

export function requestGoToPage(page: number): void {
	navigateTarget = page;
}

export function consumeNavigateTarget(): number | null {
	const target = navigateTarget;
	navigateTarget = null;
	return target;
}

export function getNavigateTarget(): number | null {
	return navigateTarget;
}
