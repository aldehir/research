import { describe, it, expect } from 'vitest';
import {
	clampScale,
	zoomIn,
	zoomOut,
	zoomByDelta,
	clampPage,
	formatZoom,
	fitToWidthScale,
	DEFAULT_SCALE,
	ZOOM_STEP,
	MIN_SCALE,
	MAX_SCALE
} from '$lib/pdf-utils';

describe('pdf-utils constants', () => {
	it('has expected defaults', () => {
		expect(DEFAULT_SCALE).toBe(1.0);
		expect(ZOOM_STEP).toBe(0.25);
		expect(MIN_SCALE).toBe(0.25);
		expect(MAX_SCALE).toBe(5.0);
	});
});

describe('clampScale', () => {
	it('returns scale unchanged when within bounds', () => {
		expect(clampScale(1.0)).toBe(1.0);
		expect(clampScale(2.5)).toBe(2.5);
	});

	it('clamps below minimum', () => {
		expect(clampScale(0.1)).toBe(MIN_SCALE);
		expect(clampScale(-1)).toBe(MIN_SCALE);
	});

	it('clamps above maximum', () => {
		expect(clampScale(6.0)).toBe(MAX_SCALE);
		expect(clampScale(100)).toBe(MAX_SCALE);
	});

	it('rounds to two decimal places', () => {
		expect(clampScale(1.333)).toBe(1.33);
		expect(clampScale(2.126)).toBe(2.13);
	});
});

describe('zoomIn', () => {
	it('increases scale by ZOOM_STEP', () => {
		expect(zoomIn(1.0)).toBe(1.25);
		expect(zoomIn(1.25)).toBe(1.5);
	});

	it('does not exceed MAX_SCALE', () => {
		expect(zoomIn(MAX_SCALE)).toBe(MAX_SCALE);
		expect(zoomIn(4.9)).toBe(MAX_SCALE);
	});
});

describe('zoomOut', () => {
	it('decreases scale by ZOOM_STEP', () => {
		expect(zoomOut(1.0)).toBe(0.75);
		expect(zoomOut(1.5)).toBe(1.25);
	});

	it('does not go below MIN_SCALE', () => {
		expect(zoomOut(MIN_SCALE)).toBe(MIN_SCALE);
		expect(zoomOut(0.3)).toBe(MIN_SCALE);
	});
});

describe('clampPage', () => {
	it('returns page unchanged when within bounds', () => {
		expect(clampPage(3, 10)).toBe(3);
		expect(clampPage(1, 1)).toBe(1);
	});

	it('clamps below 1', () => {
		expect(clampPage(0, 10)).toBe(1);
		expect(clampPage(-5, 10)).toBe(1);
	});

	it('clamps above totalPages', () => {
		expect(clampPage(15, 10)).toBe(10);
	});

	it('returns 1 when totalPages is 0', () => {
		expect(clampPage(1, 0)).toBe(1);
	});

	it('floors fractional pages', () => {
		expect(clampPage(3.7, 10)).toBe(3);
	});
});

describe('formatZoom', () => {
	it('formats scale as percentage', () => {
		expect(formatZoom(1.0)).toBe('100%');
		expect(formatZoom(1.5)).toBe('150%');
		expect(formatZoom(0.25)).toBe('25%');
	});

	it('rounds to nearest integer', () => {
		expect(formatZoom(1.333)).toBe('133%');
		expect(formatZoom(0.666)).toBe('67%');
	});
});

describe('fitToWidthScale', () => {
	it('computes scale to fit page width into container', () => {
		// page is 612px wide at scale=1, container is 800px with 16px padding
		const scale = fitToWidthScale(800, 612, 16);
		// (800 - 16) / 612 = 1.2810...
		expect(scale).toBeCloseTo(1.28, 2);
	});

	it('scales down when container is narrower than page', () => {
		const scale = fitToWidthScale(400, 612, 16);
		// (400 - 16) / 612 = 0.6274...
		expect(scale).toBeCloseTo(0.63, 2);
	});

	it('clamps to MIN_SCALE', () => {
		const scale = fitToWidthScale(50, 612, 16);
		expect(scale).toBe(MIN_SCALE);
	});

	it('clamps to MAX_SCALE', () => {
		const scale = fitToWidthScale(50000, 612, 16);
		expect(scale).toBe(MAX_SCALE);
	});

	it('uses zero padding by default', () => {
		const scale = fitToWidthScale(612, 612);
		expect(scale).toBe(1.0);
	});
});

describe('zoomByDelta', () => {
	it('zooms in with positive delta', () => {
		// deltaY -100 (scroll up) → zoom in
		const result = zoomByDelta(1.0, -100);
		expect(result).toBeGreaterThan(1.0);
	});

	it('zooms out with negative delta', () => {
		// deltaY 100 (scroll down) → zoom out
		const result = zoomByDelta(1.0, 100);
		expect(result).toBeLessThan(1.0);
	});

	it('scales proportionally to current zoom', () => {
		// Same delta at higher scale produces larger absolute change
		const changeAt1 = zoomByDelta(1.0, -100) - 1.0;
		const changeAt2 = zoomByDelta(2.0, -100) - 2.0;
		expect(changeAt2).toBeGreaterThan(changeAt1);
	});

	it('clamps to MIN_SCALE', () => {
		expect(zoomByDelta(MIN_SCALE, 10000)).toBe(MIN_SCALE);
	});

	it('clamps to MAX_SCALE', () => {
		expect(zoomByDelta(MAX_SCALE, -10000)).toBe(MAX_SCALE);
	});
});
