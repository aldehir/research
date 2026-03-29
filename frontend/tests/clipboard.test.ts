// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { writeClipboard } from '$lib/clipboard';

describe('writeClipboard', () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		document.execCommand = vi.fn().mockReturnValue(true);
	});

	it('uses navigator.clipboard.writeText when available', async () => {
		const writeText = vi.fn().mockResolvedValue(undefined);
		Object.assign(navigator, { clipboard: { writeText } });

		await writeClipboard('hello');
		expect(writeText).toHaveBeenCalledWith('hello');
	});

	it('falls back to execCommand when clipboard API is unavailable', async () => {
		Object.assign(navigator, { clipboard: undefined });

		await writeClipboard('fallback text');
		expect(document.execCommand).toHaveBeenCalledWith('copy');
	});

	it('falls back to execCommand when clipboard API rejects', async () => {
		const writeText = vi.fn().mockRejectedValue(new Error('denied'));
		Object.assign(navigator, { clipboard: { writeText } });

		await writeClipboard('retry text');
		expect(document.execCommand).toHaveBeenCalledWith('copy');
	});
});
