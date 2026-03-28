// @vitest-environment jsdom
import { describe, it, expect } from 'vitest';
import { extractPageText, extractSurroundingContext } from '$lib/text-context';

describe('extractPageText', () => {
	it('extracts text content from a div with span children', () => {
		const div = document.createElement('div');
		div.innerHTML = '<span>Hello </span><span>world</span>';
		expect(extractPageText(div)).toBe('Hello world');
	});

	it('returns empty string for empty div', () => {
		const div = document.createElement('div');
		expect(extractPageText(div)).toBe('');
	});

	it('handles nested elements', () => {
		const div = document.createElement('div');
		div.innerHTML = '<div><span>Nested</span> <span>text</span></div>';
		expect(extractPageText(div)).toBe('Nested text');
	});
});

describe('extractSurroundingContext', () => {
	const pageText = 'A'.repeat(500) + 'SELECTED_TEXT' + 'B'.repeat(500);

	it('returns window around text in the middle', () => {
		const result = extractSurroundingContext('SELECTED_TEXT', pageText, 100);
		expect(result).toContain('SELECTED_TEXT');
		expect(result.length).toBeLessThanOrEqual(100 + 'SELECTED_TEXT'.length + 100);
		expect(result.startsWith('A')).toBe(true);
		expect(result.endsWith('B')).toBe(true);
	});

	it('returns from beginning when text is at the start', () => {
		const text = 'START_HERE' + 'X'.repeat(1000);
		const result = extractSurroundingContext('START_HERE', text, 100);
		expect(result.startsWith('START_HERE')).toBe(true);
		expect(result.length).toBeLessThanOrEqual('START_HERE'.length + 100);
	});

	it('returns to end when text is at the end', () => {
		const text = 'X'.repeat(1000) + 'END_HERE';
		const result = extractSurroundingContext('END_HERE', text, 100);
		expect(result.endsWith('END_HERE')).toBe(true);
		expect(result.length).toBeLessThanOrEqual(100 + 'END_HERE'.length);
	});

	it('returns full page text when selected text is not found', () => {
		const text = 'Some page content here';
		const result = extractSurroundingContext('NOT_IN_PAGE', text);
		expect(result).toBe(text);
	});

	it('uses default context window of 500', () => {
		const result = extractSurroundingContext('SELECTED_TEXT', pageText);
		expect(result).toBe(pageText);
	});

	it('respects custom context window size', () => {
		const result = extractSurroundingContext('SELECTED_TEXT', pageText, 50);
		expect(result.length).toBeLessThanOrEqual(50 + 'SELECTED_TEXT'.length + 50);
		expect(result).toContain('SELECTED_TEXT');
	});
});
