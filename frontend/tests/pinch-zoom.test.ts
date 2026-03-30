import { describe, it, expect } from 'vitest';
import { pointerDistance, pointerMidpoint, pinchScale } from '$lib/pinch-zoom';
import { MIN_SCALE, MAX_SCALE } from '$lib/pdf-utils';

describe('pointerDistance', () => {
	it('returns 0 for identical points', () => {
		expect(pointerDistance({ x: 5, y: 5 }, { x: 5, y: 5 })).toBe(0);
	});

	it('computes horizontal distance', () => {
		expect(pointerDistance({ x: 0, y: 0 }, { x: 3, y: 0 })).toBe(3);
	});

	it('computes vertical distance', () => {
		expect(pointerDistance({ x: 0, y: 0 }, { x: 0, y: 4 })).toBe(4);
	});

	it('computes diagonal distance (3-4-5 triangle)', () => {
		expect(pointerDistance({ x: 0, y: 0 }, { x: 3, y: 4 })).toBe(5);
	});

	it('works with negative coordinates', () => {
		expect(pointerDistance({ x: -3, y: -4 }, { x: 0, y: 0 })).toBe(5);
	});
});

describe('pointerMidpoint', () => {
	it('returns the point itself for identical points', () => {
		const mid = pointerMidpoint({ x: 10, y: 20 }, { x: 10, y: 20 });
		expect(mid).toEqual({ x: 10, y: 20 });
	});

	it('computes midpoint of two points', () => {
		const mid = pointerMidpoint({ x: 0, y: 0 }, { x: 10, y: 20 });
		expect(mid).toEqual({ x: 5, y: 10 });
	});

	it('handles negative coordinates', () => {
		const mid = pointerMidpoint({ x: -10, y: -20 }, { x: 10, y: 20 });
		expect(mid).toEqual({ x: 0, y: 0 });
	});
});

describe('pinchScale', () => {
	it('returns same scale when distances are equal', () => {
		expect(pinchScale(1.0, 100, 100)).toBe(1.0);
	});

	it('scales up when fingers spread apart', () => {
		const result = pinchScale(1.0, 100, 200);
		expect(result).toBe(2.0);
	});

	it('scales down when fingers pinch together', () => {
		const result = pinchScale(2.0, 200, 100);
		expect(result).toBe(1.0);
	});

	it('clamps to MIN_SCALE', () => {
		const result = pinchScale(0.5, 200, 10);
		expect(result).toBe(MIN_SCALE);
	});

	it('clamps to MAX_SCALE', () => {
		const result = pinchScale(3.0, 100, 500);
		expect(result).toBe(MAX_SCALE);
	});

	it('returns current scale when startDistance is 0', () => {
		expect(pinchScale(1.5, 0, 200)).toBe(1.5);
	});

	it('applies ratio to current scale', () => {
		// 1.5x current scale of 2.0 = 3.0
		const result = pinchScale(2.0, 100, 150);
		expect(result).toBe(3.0);
	});
});
