import { describe, it, expect, vi, beforeEach } from 'vitest';
import { listPapers, uploadPaper, getPaper, deletePaper, getPdfUrl, updateReadingPosition, extractRegion } from '$lib/api';

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

describe('updateReadingPosition', () => {
	it('sends PATCH request with page number', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true
		}));

		await updateReadingPosition(mockPaper.id, 5);

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${mockPaper.id}/position`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ page: 5 })
		});
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'not found' })
		}));

		await expect(updateReadingPosition('nonexistent', 1)).rejects.toThrow('not found');
	});
});

describe('extractRegion', () => {
	it('sends POST with region coordinates and returns text + image', async () => {
		const mockResult = { text: 'Hello', image_data: 'base64data' };
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(mockResult)
		}));

		const result = await extractRegion(mockPaper.id, 1, 10, 20, 100, 50);

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${mockPaper.id}/region`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ page: 1, x: 10, y: 20, w: 100, h: 50 })
		});
		expect(result).toEqual(mockResult);
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'not found' })
		}));

		await expect(extractRegion('nonexistent', 1, 0, 0, 100, 100)).rejects.toThrow('not found');
	});
});

describe('getPdfUrl', () => {
	it('returns the PDF URL for a paper id', () => {
		expect(getPdfUrl(mockPaper.id)).toBe(`/api/papers/${mockPaper.id}/pdf`);
	});
});
