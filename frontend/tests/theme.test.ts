// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest';
import { getTheme, setTheme, getResolvedTheme, initTheme } from '$lib/theme.svelte';

function mockMatchMedia(prefersDark: boolean) {
	window.matchMedia = (query: string) => ({
		matches: prefersDark && query === '(prefers-color-scheme: dark)',
		media: query,
		onchange: null,
		addListener: () => {},
		removeListener: () => {},
		addEventListener: () => {},
		removeEventListener: () => {},
		dispatchEvent: () => false,
	});
}

describe('theme store', () => {
	beforeEach(() => {
		localStorage.clear();
		document.documentElement.removeAttribute('data-theme');
		mockMatchMedia(false);
	});

	describe('getTheme / setTheme', () => {
		it('defaults to system', () => {
			initTheme();
			expect(getTheme()).toBe('system');
		});

		it('persists to localStorage', () => {
			initTheme();
			setTheme('dark');
			expect(localStorage.getItem('theme')).toBe('dark');
			expect(getTheme()).toBe('dark');
		});
	});

	describe('initTheme', () => {
		it('reads theme from localStorage', () => {
			localStorage.setItem('theme', 'dark');
			initTheme();
			expect(getTheme()).toBe('dark');
			expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
		});

		it('defaults to system for invalid localStorage value', () => {
			localStorage.setItem('theme', 'garbage');
			initTheme();
			expect(getTheme()).toBe('system');
			expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
		});

		it('defaults to system when localStorage is empty', () => {
			initTheme();
			expect(getTheme()).toBe('system');
		});
	});

	describe('getResolvedTheme', () => {
		it('returns light when theme is light', () => {
			initTheme();
			setTheme('light');
			expect(getResolvedTheme()).toBe('light');
		});

		it('returns dark when theme is dark', () => {
			initTheme();
			setTheme('dark');
			expect(getResolvedTheme()).toBe('dark');
		});

		it('returns light for system when prefers-color-scheme is light', () => {
			// jsdom matchMedia returns false for all queries by default
			initTheme();
			setTheme('system');
			expect(getResolvedTheme()).toBe('light');
		});
	});
});
