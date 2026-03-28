import { describe, it, expect, vi, beforeEach } from 'vitest';
import { listPapers, uploadPaper, getPaper, deletePaper, getPdfUrl } from '$lib/api';

const mockPaper = {
	id: '123e4567-e89b-12d3-a456-426614174000',
	title: 'Test Paper',
	file_path: '/papers/test.pdf',
	file_size: 1024,
	created_at: '2026-01-01T00:00:00Z'
};

beforeEach(() => {
	vi.restoreAllMocks();
});

describe('listPapers', () => {
	it('fetches and returns parsed array', async () => {
		const papers = [mockPaper];
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(papers)
		}));

		const result = await listPapers();

		expect(fetch).toHaveBeenCalledWith('/api/papers');
		expect(result).toEqual(papers);
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'server error' })
		}));

		await expect(listPapers()).rejects.toThrow('server error');
	});
});

describe('uploadPaper', () => {
	it('sends FormData with file', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(mockPaper)
		}));

		const file = new File(['pdf content'], 'test.pdf', { type: 'application/pdf' });
		const result = await uploadPaper(file);

		expect(fetch).toHaveBeenCalledWith('/api/papers', {
			method: 'POST',
			body: expect.any(FormData)
		});

		const callArgs = vi.mocked(fetch).mock.calls[0];
		const formData = callArgs[1]?.body as FormData;
		expect(formData.get('file')).toBe(file);
		expect(result).toEqual(mockPaper);
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'invalid file' })
		}));

		const file = new File([''], 'bad.pdf', { type: 'application/pdf' });
		await expect(uploadPaper(file)).rejects.toThrow('invalid file');
	});
});

describe('getPaper', () => {
	it('fetches a single paper by id', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(mockPaper)
		}));

		const result = await getPaper(mockPaper.id);

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${mockPaper.id}`);
		expect(result).toEqual(mockPaper);
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'not found' })
		}));

		await expect(getPaper('nonexistent')).rejects.toThrow('not found');
	});
});

describe('deletePaper', () => {
	it('sends DELETE request', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true
		}));

		await deletePaper(mockPaper.id);

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${mockPaper.id}`, {
			method: 'DELETE'
		});
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'forbidden' })
		}));

		await expect(deletePaper(mockPaper.id)).rejects.toThrow('forbidden');
	});
});

describe('getPdfUrl', () => {
	it('returns the PDF URL for a paper id', () => {
		expect(getPdfUrl(mockPaper.id)).toBe(`/api/papers/${mockPaper.id}/pdf`);
	});
});
