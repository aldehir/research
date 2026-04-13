import { describe, it, expect, vi, beforeEach } from 'vitest';
import { resizeImage } from '../src/lib/resize-image';

describe('resizeImage', () => {
	let mockCanvas: { width: number; height: number; toDataURL: ReturnType<typeof vi.fn> };
	let mockCtx: { drawImage: ReturnType<typeof vi.fn> };

	beforeEach(() => {
		mockCtx = { drawImage: vi.fn() };
		mockCanvas = {
			width: 0,
			height: 0,
			toDataURL: vi.fn(() => 'data:image/png;base64,AAAA'),
		};
		vi.stubGlobal('document', {
			createElement: vi.fn(() => mockCanvas),
		});
		(mockCanvas as unknown as HTMLCanvasElement).getContext = vi.fn(() => mockCtx);
	});

	function makeBitmap(w: number, h: number): ImageBitmap {
		return { width: w, height: h, close: vi.fn() } as unknown as ImageBitmap;
	}

	it('returns base64 without data: prefix', async () => {
		vi.stubGlobal('createImageBitmap', vi.fn(() => Promise.resolve(makeBitmap(100, 50))));
		const blob = new Blob(['x'], { type: 'image/png' });
		const result = await resizeImage(blob);
		expect(result).toBe('AAAA');
		expect(result).not.toContain('data:');
	});

	it('does not resize images within the limit', async () => {
		vi.stubGlobal('createImageBitmap', vi.fn(() => Promise.resolve(makeBitmap(1024, 768))));
		const blob = new Blob(['x'], { type: 'image/png' });
		await resizeImage(blob);
		expect(mockCanvas.width).toBe(1024);
		expect(mockCanvas.height).toBe(768);
	});

	it('scales down landscape images exceeding max dimension', async () => {
		vi.stubGlobal('createImageBitmap', vi.fn(() => Promise.resolve(makeBitmap(4096, 2048))));
		const blob = new Blob(['x'], { type: 'image/png' });
		await resizeImage(blob);
		expect(mockCanvas.width).toBe(2048);
		expect(mockCanvas.height).toBe(1024);
	});

	it('scales down portrait images exceeding max dimension', async () => {
		vi.stubGlobal('createImageBitmap', vi.fn(() => Promise.resolve(makeBitmap(1000, 4000))));
		const blob = new Blob(['x'], { type: 'image/png' });
		await resizeImage(blob);
		expect(mockCanvas.width).toBe(512);
		expect(mockCanvas.height).toBe(2048);
	});

	it('closes the ImageBitmap after use', async () => {
		const bmp = makeBitmap(100, 100);
		vi.stubGlobal('createImageBitmap', vi.fn(() => Promise.resolve(bmp)));
		const blob = new Blob(['x'], { type: 'image/png' });
		await resizeImage(blob);
		expect(bmp.close).toHaveBeenCalled();
	});
});
