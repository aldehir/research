// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest';
import { getTheme, setTheme, getResolvedTheme, initTheme } from '$lib/theme.svelte';

describe('theme store', () => {
	beforeEach(() => {
		localStorage.clear();
		document.documentElement.removeAttribute('data-theme');
	});

	describe('getTheme / setTheme', () => {
		it('defaults to light', () => {
			initTheme();
			expect(getTheme()).toBe('light');
		});

		it('persists to localStorage', () => {
			initTheme();
			setTheme('dark');
			expect(localStorage.getItem('theme')).toBe('dark');
			expect(getTheme()).toBe('dark');
		});

		it('sets data-theme attribute on html element', () => {
			initTheme();
			setTheme('dark');
			expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
			setTheme('light');
			expect(document.documentElement.getAttribute('data-theme')).toBe('light');
		});
	});

	describe('initTheme', () => {
		it('reads theme from localStorage', () => {
			localStorage.setItem('theme', 'dark');
			initTheme();
			expect(getTheme()).toBe('dark');
			expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
		});

		it('defaults to light for invalid localStorage value', () => {
			localStorage.setItem('theme', 'garbage');
			initTheme();
			expect(getTheme()).toBe('light');
			expect(document.documentElement.getAttribute('data-theme')).toBe('light');
		});

		it('migrates legacy system value to light', () => {
			localStorage.setItem('theme', 'system');
			initTheme();
			expect(getTheme()).toBe('light');
			expect(document.documentElement.getAttribute('data-theme')).toBe('light');
		});

		it('defaults to light when localStorage is empty', () => {
			initTheme();
			expect(getTheme()).toBe('light');
			expect(document.documentElement.getAttribute('data-theme')).toBe('light');
		});
	});

	describe('toggleTheme', () => {
		it('toggles from light to dark on a single call', async () => {
			const { toggleTheme } = await import('$lib/theme.svelte');
			initTheme();
			expect(getTheme()).toBe('light');
			toggleTheme();
			expect(getTheme()).toBe('dark');
			expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
		});

		it('toggles from dark to light on a single call', async () => {
			const { toggleTheme } = await import('$lib/theme.svelte');
			initTheme();
			setTheme('dark');
			expect(getTheme()).toBe('dark');
			toggleTheme();
			expect(getTheme()).toBe('light');
			expect(document.documentElement.getAttribute('data-theme')).toBe('light');
		});

		it('round-trips correctly', async () => {
			const { toggleTheme } = await import('$lib/theme.svelte');
			initTheme();
			expect(getTheme()).toBe('light');
			toggleTheme();
			expect(getTheme()).toBe('dark');
			toggleTheme();
			expect(getTheme()).toBe('light');
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
	});
});
