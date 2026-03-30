import { clampScale } from './pdf-utils';

export interface Point {
	x: number;
	y: number;
}

export function pointerDistance(a: Point, b: Point): number {
	const dx = b.x - a.x;
	const dy = b.y - a.y;
	return Math.sqrt(dx * dx + dy * dy);
}

export function pointerMidpoint(a: Point, b: Point): Point {
	return {
		x: (a.x + b.x) / 2,
		y: (a.y + b.y) / 2
	};
}

export function pinchScale(currentScale: number, startDistance: number, currentDistance: number): number {
	if (startDistance === 0) return currentScale;
	const ratio = currentDistance / startDistance;
	return clampScale(currentScale * ratio);
}
