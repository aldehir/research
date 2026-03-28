import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as api from '$lib/api';
import {
	getPapers,
	getSelectedPaper,
	loadPapers,
	selectPaper,
	upload,
	remove
} from '$lib/papers.svelte';

vi.mock('$lib/api', () => ({
	listPapers: vi.fn(),
	uploadPaper: vi.fn(),
	deletePaper: vi.fn()
}));

const mockPaper: api.Paper = {
	id: '123e4567-e89b-12d3-a456-426614174000',
	title: 'Test Paper',
	file_path: '/papers/test.pdf',
	file_size: 1024,
	created_at: '2026-01-01T00:00:00Z'
};

const mockPaper2: api.Paper = {
	id: '223e4567-e89b-12d3-a456-426614174000',
	title: 'Another Paper',
	file_path: '/papers/another.pdf',
	file_size: 2048,
	created_at: '2026-01-02T00:00:00Z'
};

beforeEach(() => {
	vi.restoreAllMocks();
	// Reset store state by loading empty
	vi.mocked(api.listPapers).mockResolvedValue([]);
});

describe('paper store', () => {
	it('starts with empty papers', () => {
		expect(getPapers()).toEqual([]);
	});

	it('loads papers from API', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);

		await loadPapers();

		expect(api.listPapers).toHaveBeenCalled();
		expect(getPapers()).toEqual([mockPaper, mockPaper2]);
	});

	it('selects a paper by id', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);
		await loadPapers();

		selectPaper(mockPaper2.id);

		expect(getSelectedPaper()).toEqual(mockPaper2);
	});

	it('returns null when no paper is selected', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper]);
		await loadPapers();

		expect(getSelectedPaper()).toBeNull();
	});

	it('uploads a paper and refreshes list', async () => {
		const file = new File(['pdf'], 'test.pdf', { type: 'application/pdf' });
		vi.mocked(api.uploadPaper).mockResolvedValue(mockPaper);
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper]);

		await upload(file);

		expect(api.uploadPaper).toHaveBeenCalledWith(file);
		expect(api.listPapers).toHaveBeenCalled();
		expect(getPapers()).toEqual([mockPaper]);
	});

	it('removes a paper and refreshes list', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);
		await loadPapers();

		vi.mocked(api.deletePaper).mockResolvedValue(undefined);
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper2]);

		await remove(mockPaper.id);

		expect(api.deletePaper).toHaveBeenCalledWith(mockPaper.id);
		expect(getPapers()).toEqual([mockPaper2]);
	});

	it('clears selection when selected paper is removed', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);
		await loadPapers();
		selectPaper(mockPaper.id);
		expect(getSelectedPaper()).toEqual(mockPaper);

		vi.mocked(api.deletePaper).mockResolvedValue(undefined);
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper2]);

		await remove(mockPaper.id);

		expect(getSelectedPaper()).toBeNull();
	});
});
