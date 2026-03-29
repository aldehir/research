/**
 * Keyboard navigation helpers for the PDF viewer.
 * Pure functions — no DOM dependencies, easy to test.
 */

/** Scroll distance in pixels for arrow key navigation. */
export const SMALL_JUMP = 100;

/**
 * Return the scroll delta (in pixels) for a navigation key press,
 * or null if the key is not a navigation key we handle.
 */
export function getScrollDelta(key: string, containerHeight: number): number | null {
	const halfPage = containerHeight / 2;
	switch (key) {
		case ' ':
		case 'PageDown':
			return halfPage;
		case 'PageUp':
			return -halfPage;
		case 'ArrowDown':
			return SMALL_JUMP;
		case 'ArrowUp':
			return -SMALL_JUMP;
		default:
			return null;
	}
}

/**
 * Whether the keydown handler should be skipped because focus
 * is in a text-entry element.
 */
export function shouldSkipKeyHandler(target: HTMLElement): boolean {
	const tag = target.tagName;
	return tag === 'INPUT' || tag === 'TEXTAREA';
}
