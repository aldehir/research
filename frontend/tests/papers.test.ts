import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as api from '$lib/api';
import { papersStore } from '$lib/papers.svelte';

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
		expect(papersStore.papers).toEqual([]);
	});

	it('loads papers from API', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);

		await papersStore.load();

		expect(api.listPapers).toHaveBeenCalled();
		expect(papersStore.papers).toEqual([mockPaper, mockPaper2]);
	});

	it('selects a paper by id', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);
		await papersStore.load();

		papersStore.select(mockPaper2.id);

		expect(papersStore.selectedPaper).toEqual(mockPaper2);
	});

	it('returns null when no paper is selected', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper]);
		await papersStore.load();

		expect(papersStore.selectedPaper).toBeNull();
	});

	it('uploads a paper and refreshes list', async () => {
		const file = new File(['pdf'], 'test.pdf', { type: 'application/pdf' });
		vi.mocked(api.uploadPaper).mockResolvedValue(mockPaper);
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper]);

		await papersStore.upload(file);

		expect(api.uploadPaper).toHaveBeenCalledWith(file);
		expect(api.listPapers).toHaveBeenCalled();
		expect(papersStore.papers).toEqual([mockPaper]);
	});

	it('removes a paper and refreshes list', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);
		await papersStore.load();

		vi.mocked(api.deletePaper).mockResolvedValue(undefined);
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper2]);

		await papersStore.remove(mockPaper.id);

		expect(api.deletePaper).toHaveBeenCalledWith(mockPaper.id);
		expect(papersStore.papers).toEqual([mockPaper2]);
	});

	it('sets loading true while fetching and false after', async () => {
		let resolveFetch: (papers: api.Paper[]) => void;
		vi.mocked(api.listPapers).mockImplementation(
			() => new Promise(resolve => { resolveFetch = resolve; })
		);

		expect(papersStore.loading).toBe(false);

		const loadPromise = papersStore.load();
		expect(papersStore.loading).toBe(true);

		resolveFetch!([mockPaper]);
		await loadPromise;
		expect(papersStore.loading).toBe(false);
		expect(papersStore.papers).toEqual([mockPaper]);
	});

	it('sets loading false on fetch error', async () => {
		vi.mocked(api.listPapers).mockRejectedValue(new Error('network'));

		await papersStore.load().catch(() => {});
		expect(papersStore.loading).toBe(false);
	});

	it('clears selection when selected paper is removed', async () => {
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);
		await papersStore.load();
		papersStore.select(mockPaper.id);
		expect(papersStore.selectedPaper).toEqual(mockPaper);

		vi.mocked(api.deletePaper).mockResolvedValue(undefined);
		vi.mocked(api.listPapers).mockResolvedValue([mockPaper2]);

		await papersStore.remove(mockPaper.id);

		expect(papersStore.selectedPaper).toBeNull();
	});
});
