import { describe, it, expect } from 'vitest';
import { findActiveTocEntry, type TocEntry } from '$lib/pdf-outline';

describe('findActiveTocEntry', () => {
	const flat: TocEntry[] = [
		{ title: 'Chapter 1', pageNumber: 1, children: [] },
		{ title: 'Chapter 2', pageNumber: 5, children: [] },
		{ title: 'Chapter 3', pageNumber: 10, children: [] }
	];

	it('returns null for empty entries', () => {
		expect(findActiveTocEntry([], 1)).toBeNull();
	});

	it('returns the entry whose pageNumber matches current page', () => {
		expect(findActiveTocEntry(flat, 5)).toBe(flat[1]);
	});

	it('returns the last entry whose pageNumber <= current page', () => {
		// Page 7 is between Chapter 2 (p5) and Chapter 3 (p10)
		expect(findActiveTocEntry(flat, 7)).toBe(flat[1]);
	});

	it('returns first entry when current page is before any entry', () => {
		const entries: TocEntry[] = [
			{ title: 'A', pageNumber: 3, children: [] },
			{ title: 'B', pageNumber: 6, children: [] }
		];
		expect(findActiveTocEntry(entries, 1)).toBe(entries[0]);
	});

	it('returns last entry when current page is past all entries', () => {
		expect(findActiveTocEntry(flat, 100)).toBe(flat[2]);
	});

	it('prefers deepest nested match', () => {
		const nested: TocEntry[] = [
			{
				title: 'Part 1',
				pageNumber: 1,
				children: [
					{ title: 'Section 1.1', pageNumber: 2, children: [] },
					{ title: 'Section 1.2', pageNumber: 5, children: [] }
				]
			},
			{ title: 'Part 2', pageNumber: 10, children: [] }
		];
		// Page 3 should match Section 1.1 (deepest entry with pageNumber <= 3)
		const result = findActiveTocEntry(nested, 3);
		expect(result?.title).toBe('Section 1.1');
	});

	it('matches parent when page is before first child', () => {
		const nested: TocEntry[] = [
			{
				title: 'Part 1',
				pageNumber: 1,
				children: [
					{ title: 'Section 1.1', pageNumber: 5, children: [] }
				]
			}
		];
		// Page 3 — Part 1 starts at 1, but Section 1.1 starts at 5, so Part 1 is the match
		const result = findActiveTocEntry(nested, 3);
		expect(result?.title).toBe('Part 1');
	});
});
