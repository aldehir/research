// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

const css = readFileSync(resolve(__dirname, '../src/lib/theme.css'), 'utf-8');
const html = readFileSync(resolve(__dirname, '../src/app.html'), 'utf-8');

describe('overscroll-behavior CSS', () => {
	it('theme.css sets overscroll-behavior: none on html and body', () => {
		expect(css).toContain('overscroll-behavior: none');
	});

	it('theme.css sets overflow: hidden on html', () => {
		const htmlBlock = css.match(/html\s*\{[^}]*\}/s);
		expect(htmlBlock).not.toBeNull();
		expect(htmlBlock![0]).toContain('overflow: hidden');
	});
});

describe('viewport and PWA meta', () => {
	it('includes viewport-fit=cover for safe area support', () => {
		expect(html).toContain('viewport-fit=cover');
	});

	it('includes maximum-scale=1 to prevent native page zoom', () => {
		expect(html).toContain('maximum-scale=1');
	});

	it('includes apple-mobile-web-app-capable meta', () => {
		expect(html).toContain('apple-mobile-web-app-capable');
	});

	it('links to manifest.json', () => {
		expect(html).toContain('rel="manifest"');
		expect(html).toContain('manifest.json');
	});
});

describe('manifest.json', () => {
	const manifest = JSON.parse(
		readFileSync(resolve(__dirname, '../static/manifest.json'), 'utf-8')
	);

	it('sets display to standalone', () => {
		expect(manifest.display).toBe('standalone');
	});

	it('has a start_url', () => {
		expect(manifest.start_url).toBe('/');
	});
});

describe('fullscreen store', () => {
	beforeEach(() => {
		// Mock Fullscreen API
		Object.defineProperty(document, 'fullscreenElement', {
			writable: true,
			value: null
		});
		document.exitFullscreen = vi.fn().mockResolvedValue(undefined);
		document.documentElement.requestFullscreen = vi.fn().mockResolvedValue(undefined);
	});

	it('defaults to false when no fullscreen element', async () => {
		const { isFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		initFullscreen();
		expect(isFullscreen()).toBe(false);
	});

	it('calls requestFullscreen on toggle when not fullscreen', async () => {
		const { toggleFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		initFullscreen();
		toggleFullscreen();
		expect(document.documentElement.requestFullscreen).toHaveBeenCalled();
	});

	it('calls exitFullscreen on toggle when fullscreen', async () => {
		const { toggleFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		(document as unknown as Record<string, unknown>).fullscreenElement = document.documentElement;
		initFullscreen();
		toggleFullscreen();
		expect(document.exitFullscreen).toHaveBeenCalled();
	});

	it('tracks fullscreenchange events', async () => {
		const { isFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		initFullscreen();
		expect(isFullscreen()).toBe(false);

		(document as unknown as Record<string, unknown>).fullscreenElement = document.documentElement;
		document.dispatchEvent(new Event('fullscreenchange'));
		expect(isFullscreen()).toBe(true);

		(document as unknown as Record<string, unknown>).fullscreenElement = null;
		document.dispatchEvent(new Event('fullscreenchange'));
		expect(isFullscreen()).toBe(false);
	});
});
