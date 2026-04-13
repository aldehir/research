import { describe, it, expect } from 'vitest';
import { formatToolLabel, formatToolArgs } from '$lib/tool-display';

describe('formatToolLabel', () => {
	it('returns label for search_pdf', () => {
		expect(formatToolLabel('search_pdf')).toBe('Searched PDF');
	});

	it('returns label for read_page', () => {
		expect(formatToolLabel('read_page')).toBe('Read page');
	});

	it('returns label for go_to_page', () => {
		expect(formatToolLabel('go_to_page')).toBe('Navigated to page');
	});

	it('returns label for snapshot_page', () => {
		expect(formatToolLabel('snapshot_page')).toBe('Page snapshot');
	});

	it('returns generic label for unknown tools', () => {
		expect(formatToolLabel('unknown_tool')).toBe('Used tool');
	});
});

describe('formatToolArgs', () => {
	it('formats search_pdf query', () => {
		expect(formatToolArgs('search_pdf', { query: 'attention mechanism' })).toBe('"attention mechanism"');
	});

	it('formats read_page page number', () => {
		expect(formatToolArgs('read_page', { page: 5 })).toBe('page 5');
	});

	it('formats go_to_page page number', () => {
		expect(formatToolArgs('go_to_page', { page: 3 })).toBe('page 3');
	});

	it('formats snapshot_page page number', () => {
		expect(formatToolArgs('snapshot_page', { page: 2 })).toBe('page 2');
	});

	it('falls back to JSON for unknown tools', () => {
		expect(formatToolArgs('other', { foo: 'bar' })).toBe('{"foo":"bar"}');
	});
});
