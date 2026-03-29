export const DEFAULT_SCALE = 1.0;
export const ZOOM_STEP = 0.25;
export const MIN_SCALE = 0.25;
export const MAX_SCALE = 5.0;

export function clampScale(scale: number): number {
	if (scale < MIN_SCALE) return MIN_SCALE;
	if (scale > MAX_SCALE) return MAX_SCALE;
	return Math.round(scale * 100) / 100;
}

export function zoomIn(scale: number): number {
	return clampScale(scale + ZOOM_STEP);
}

export function zoomOut(scale: number): number {
	return clampScale(scale - ZOOM_STEP);
}

export function clampPage(page: number, totalPages: number): number {
	if (totalPages <= 0) return 1;
	if (page < 1) return 1;
	if (page > totalPages) return totalPages;
	return Math.floor(page);
}

export function formatZoom(scale: number): string {
	return `${Math.round(scale * 100)}%`;
}

export function zoomByDelta(scale: number, deltaY: number): number {
	const factor = 1 - deltaY * 0.002;
	return clampScale(scale * factor);
}

export function maxPageWidth(widths: number[]): number {
	if (widths.length === 0) return 0;
	return Math.max(...widths);
}

export function fitToWidthScale(
	containerWidth: number,
	pageWidth: number,
	padding: number = 0
): number {
	return clampScale((containerWidth - padding) / pageWidth);
}
