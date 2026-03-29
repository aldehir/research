// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest';
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

describe('viewport meta', () => {
	it('includes viewport-fit=cover for safe area support', () => {
		expect(html).toContain('viewport-fit=cover');
	});
});

describe('fullscreen store', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	it('defaults to false', async () => {
		const { isFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		initFullscreen();
		expect(isFullscreen()).toBe(false);
	});

	it('toggles fullscreen on and off', async () => {
		const { isFullscreen, toggleFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		initFullscreen();
		expect(isFullscreen()).toBe(false);
		toggleFullscreen();
		expect(isFullscreen()).toBe(true);
		toggleFullscreen();
		expect(isFullscreen()).toBe(false);
	});

	it('persists to localStorage', async () => {
		const { toggleFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		initFullscreen();
		toggleFullscreen();
		expect(localStorage.getItem('fullscreen')).toBe('true');
		toggleFullscreen();
		expect(localStorage.getItem('fullscreen')).toBe('false');
	});

	it('restores from localStorage on init', async () => {
		const { isFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		localStorage.setItem('fullscreen', 'true');
		initFullscreen();
		expect(isFullscreen()).toBe(true);
	});

	it('sets data-fullscreen attribute on html element', async () => {
		const { toggleFullscreen, initFullscreen } = await import('$lib/fullscreen.svelte');
		initFullscreen();
		toggleFullscreen();
		expect(document.documentElement.hasAttribute('data-fullscreen')).toBe(true);
		toggleFullscreen();
		expect(document.documentElement.hasAttribute('data-fullscreen')).toBe(false);
	});
});
