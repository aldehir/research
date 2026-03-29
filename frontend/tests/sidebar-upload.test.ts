import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as api from '$lib/api';
import { papersStore } from '$lib/papers.svelte';

vi.mock('$lib/api', () => ({
	listPapers: vi.fn(),
	uploadPaper: vi.fn(),
	deletePaper: vi.fn()
}));

beforeEach(() => {
	vi.restoreAllMocks();
	vi.mocked(api.listPapers).mockResolvedValue([]);
});

/**
 * Test the handleFile logic that will be inlined in +page.svelte sidebar.
 * We extract and test the pure logic: PDF validation, upload call, error handling.
 */

// Extracted upload handler logic (mirrors what will be in +page.svelte)
async function handleSidebarUpload(
	file: File,
	callbacks: {
		setUploading: (v: boolean) => void;
		setError: (v: string | null) => void;
	}
) {
	if (!file.name.toLowerCase().endsWith('.pdf')) {
		callbacks.setError('Only PDF files are accepted');
		return;
	}
	callbacks.setError(null);
	callbacks.setUploading(true);
	try {
		await papersStore.upload(file);
	} catch (e) {
		callbacks.setError(e instanceof Error ? e.message : 'Upload failed');
	} finally {
		callbacks.setUploading(false);
	}
}

describe('sidebar upload handler', () => {
	it('rejects non-PDF files with error message', async () => {
		const setUploading = vi.fn();
		const setError = vi.fn();
		const file = new File(['data'], 'readme.txt', { type: 'text/plain' });

		await handleSidebarUpload(file, { setUploading, setError });

		expect(setError).toHaveBeenCalledWith('Only PDF files are accepted');
		expect(setUploading).not.toHaveBeenCalled();
		expect(api.uploadPaper).not.toHaveBeenCalled();
	});

	it('accepts PDF files and calls upload', async () => {
		const setUploading = vi.fn();
		const setError = vi.fn();
		const file = new File(['pdf'], 'paper.pdf', { type: 'application/pdf' });
		vi.mocked(api.uploadPaper).mockResolvedValue({
			id: 'abc', title: 'paper', file_path: '/p.pdf', file_size: 100, created_at: ''
		});
		vi.mocked(api.listPapers).mockResolvedValue([]);

		await handleSidebarUpload(file, { setUploading, setError });

		expect(setError).toHaveBeenCalledWith(null);
		expect(setUploading).toHaveBeenCalledWith(true);
		expect(setUploading).toHaveBeenCalledWith(false);
		expect(api.uploadPaper).toHaveBeenCalledWith(file);
	});

	it('accepts .PDF (case insensitive)', async () => {
		const setUploading = vi.fn();
		const setError = vi.fn();
		const file = new File(['pdf'], 'THESIS.PDF', { type: 'application/pdf' });
		vi.mocked(api.uploadPaper).mockResolvedValue({
			id: 'abc', title: 'thesis', file_path: '/t.pdf', file_size: 100, created_at: ''
		});
		vi.mocked(api.listPapers).mockResolvedValue([]);

		await handleSidebarUpload(file, { setUploading, setError });

		expect(api.uploadPaper).toHaveBeenCalledWith(file);
	});

	it('sets error on upload failure', async () => {
		const setUploading = vi.fn();
		const setError = vi.fn();
		const file = new File(['pdf'], 'paper.pdf', { type: 'application/pdf' });
		vi.mocked(api.uploadPaper).mockRejectedValue(new Error('Server error'));

		await handleSidebarUpload(file, { setUploading, setError });

		expect(setError).toHaveBeenCalledWith('Server error');
		expect(setUploading).toHaveBeenLastCalledWith(false);
	});

	it('sets generic error for non-Error throws', async () => {
		const setUploading = vi.fn();
		const setError = vi.fn();
		const file = new File(['pdf'], 'paper.pdf', { type: 'application/pdf' });
		vi.mocked(api.uploadPaper).mockRejectedValue('unknown');

		await handleSidebarUpload(file, { setUploading, setError });

		expect(setError).toHaveBeenCalledWith('Upload failed');
	});
});

describe('drag-drop file extraction', () => {
	it('extracts first file from DataTransfer', () => {
		// Simulates the logic: event.dataTransfer?.files[0]
		const file = new File(['pdf'], 'paper.pdf', { type: 'application/pdf' });
		const files = [file] as unknown as FileList;
		const dataTransfer = { files } as DataTransfer;

		const extracted = dataTransfer.files[0];
		expect(extracted).toBe(file);
		expect(extracted.name).toBe('paper.pdf');
	});

	it('handles empty DataTransfer gracefully', () => {
		const files = [] as unknown as FileList;
		const dataTransfer = { files } as DataTransfer;

		const extracted = dataTransfer.files[0];
		expect(extracted).toBeUndefined();
	});
});
