import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as api from '$lib/api';
import { papersStore } from '$lib/papers.svelte';

vi.mock('$lib/api', () => ({
	listPapers: vi.fn(),
	uploadPaper: vi.fn(),
	deletePaper: vi.fn(),
	getPaper: vi.fn()
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

beforeEach(async () => {
	vi.restoreAllMocks();
	vi.mocked(api.listPapers).mockResolvedValue([]);
	await papersStore.load();
	papersStore.deselect();
});

describe('paper routing', () => {
	describe('navigating to /papers/<id> loads the paper', () => {
		it('loadAndSelect fetches papers and sets selectedPaper', async () => {
			vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);

			await papersStore.loadAndSelect(mockPaper.id);

			expect(papersStore.selectedId).toBe(mockPaper.id);
			expect(papersStore.selectedPaper).toEqual(mockPaper);
		});

		it('loadAndSelect reuses already-loaded papers', async () => {
			vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);
			await papersStore.load();
			vi.mocked(api.listPapers).mockClear();

			await papersStore.loadAndSelect(mockPaper2.id);

			expect(api.listPapers).not.toHaveBeenCalled();
			expect(papersStore.selectedPaper).toEqual(mockPaper2);
		});
	});

	describe('selecting a paper updates the URL', () => {
		it('select sets selectedId for navigation', async () => {
			vi.mocked(api.listPapers).mockResolvedValue([mockPaper]);
			await papersStore.load();

			papersStore.select(mockPaper.id);

			expect(papersStore.selectedId).toBe(mockPaper.id);
			expect(papersStore.selectedPaper).toEqual(mockPaper);
		});
	});

	describe('browser back clears selection', () => {
		it('deselect clears selectedId and selectedPaper', async () => {
			vi.mocked(api.listPapers).mockResolvedValue([mockPaper]);
			await papersStore.load();
			papersStore.select(mockPaper.id);
			expect(papersStore.selectedPaper).toEqual(mockPaper);

			papersStore.deselect();

			expect(papersStore.selectedId).toBeNull();
			expect(papersStore.selectedPaper).toBeNull();
		});

		it('removing the selected paper clears selection', async () => {
			vi.mocked(api.listPapers).mockResolvedValue([mockPaper, mockPaper2]);
			await papersStore.load();
			papersStore.select(mockPaper.id);

			vi.mocked(api.deletePaper).mockResolvedValue(undefined);
			vi.mocked(api.listPapers).mockResolvedValue([mockPaper2]);

			await papersStore.remove(mockPaper.id);

			expect(papersStore.selectedId).toBeNull();
			expect(papersStore.selectedPaper).toBeNull();
		});
	});
});
