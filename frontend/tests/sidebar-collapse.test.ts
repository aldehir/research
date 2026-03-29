import { describe, it, expect, beforeEach } from 'vitest';
import { PanelLeftOpen, PanelLeftClose } from '$lib/icons';
import {
	isSidebarCollapsed,
	toggleSidebarCollapsed,
	setSidebarCollapsed,
	setSidebarWidth,
	getSidebarWidth
} from '$lib/panel-widths.svelte';
import { SIDEBAR_DEFAULT } from '$lib/panel-resize';

describe('sidebar collapse icons', () => {
	it('exports PanelLeftOpen as a non-empty SVG path string', () => {
		expect(typeof PanelLeftOpen).toBe('string');
		expect(PanelLeftOpen.length).toBeGreaterThan(0);
		expect(PanelLeftOpen).toContain('<');
	});

	it('exports PanelLeftClose as a non-empty SVG path string', () => {
		expect(typeof PanelLeftClose).toBe('string');
		expect(PanelLeftClose.length).toBeGreaterThan(0);
		expect(PanelLeftClose).toContain('<');
	});
});

describe('sidebar collapse state', () => {
	beforeEach(() => {
		setSidebarCollapsed(false);
		setSidebarWidth(SIDEBAR_DEFAULT);
	});

	it('starts expanded', () => {
		expect(isSidebarCollapsed()).toBe(false);
	});

	it('toggles to collapsed', () => {
		toggleSidebarCollapsed();
		expect(isSidebarCollapsed()).toBe(true);
	});

	it('toggles back to expanded', () => {
		toggleSidebarCollapsed();
		toggleSidebarCollapsed();
		expect(isSidebarCollapsed()).toBe(false);
	});

	it('restores previous sidebar width after expand', () => {
		setSidebarWidth(400);
		toggleSidebarCollapsed();
		expect(isSidebarCollapsed()).toBe(true);
		toggleSidebarCollapsed();
		expect(isSidebarCollapsed()).toBe(false);
		expect(getSidebarWidth()).toBe(400);
	});
});
