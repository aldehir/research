// @vitest-environment jsdom
import { describe, it, expect } from 'vitest';
import { getScrollDelta, shouldSkipKeyHandler, SMALL_JUMP } from '$lib/pdf-keys';

const CONTAINER_HEIGHT = 800;
const HALF_PAGE = CONTAINER_HEIGHT / 2;

describe('getScrollDelta', () => {
	it('Space scrolls down by half page', () => {
		expect(getScrollDelta(' ', CONTAINER_HEIGHT)).toBe(HALF_PAGE);
	});

	it('PageDown scrolls down by half page', () => {
		expect(getScrollDelta('PageDown', CONTAINER_HEIGHT)).toBe(HALF_PAGE);
	});

	it('PageUp scrolls up by half page', () => {
		expect(getScrollDelta('PageUp', CONTAINER_HEIGHT)).toBe(-HALF_PAGE);
	});

	it('ArrowDown scrolls down by small increment', () => {
		expect(getScrollDelta('ArrowDown', CONTAINER_HEIGHT)).toBe(SMALL_JUMP);
	});

	it('ArrowUp scrolls up by small increment', () => {
		expect(getScrollDelta('ArrowUp', CONTAINER_HEIGHT)).toBe(-SMALL_JUMP);
	});

	it('returns null for unhandled keys', () => {
		expect(getScrollDelta('a', CONTAINER_HEIGHT)).toBeNull();
		expect(getScrollDelta('Enter', CONTAINER_HEIGHT)).toBeNull();
		expect(getScrollDelta('Tab', CONTAINER_HEIGHT)).toBeNull();
	});

	it('scales half-page delta with container height', () => {
		expect(getScrollDelta(' ', 600)).toBe(300);
		expect(getScrollDelta('PageDown', 1200)).toBe(600);
		expect(getScrollDelta('PageUp', 1000)).toBe(-500);
	});
});

describe('shouldSkipKeyHandler', () => {
	it('returns true for INPUT elements', () => {
		const input = document.createElement('input');
		expect(shouldSkipKeyHandler(input)).toBe(true);
	});

	it('returns true for TEXTAREA elements', () => {
		const textarea = document.createElement('textarea');
		expect(shouldSkipKeyHandler(textarea)).toBe(true);
	});

	it('returns false for DIV elements', () => {
		const div = document.createElement('div');
		expect(shouldSkipKeyHandler(div)).toBe(false);
	});

	it('returns false for BUTTON elements', () => {
		const button = document.createElement('button');
		expect(shouldSkipKeyHandler(button)).toBe(false);
	});
});
