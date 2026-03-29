// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest';
import {
	getCurrentPage,
	setCurrentPage
} from '$lib/pdf-context.svelte';

describe('pdf-context', () => {
	beforeEach(() => {
		setCurrentPage(1);
	});

	it('stores and retrieves current page', () => {
		setCurrentPage(5);
		expect(getCurrentPage()).toBe(5);
	});

	it('defaults to page 1', () => {
		expect(getCurrentPage()).toBe(1);
	});
});
