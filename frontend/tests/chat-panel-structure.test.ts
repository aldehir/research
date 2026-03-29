import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

/**
 * Structural tests for ChatPanel.svelte — ensures the dropdown is nested
 * inside a position:relative wrapper so it anchors below the trigger button.
 */
describe('ChatPanel dropdown structure', () => {
	const fullSrc = readFileSync(
		resolve(__dirname, '../src/lib/ChatPanel.svelte'),
		'utf-8'
	);
	// Only look at the template, not <style> or <script>
	const styleStart = fullSrc.indexOf('<style>');
	const src = fullSrc.slice(0, styleStart);

	it('dropdown is inside the session-picker container, not a sibling', () => {
		// The dropdown should appear between session-picker and chat-header closing.
		// Find positions relative to session-picker in the template.
		const pickerStart = src.indexOf('class="session-picker"');
		const dropdownPos = src.indexOf('class="dropdown"', pickerStart);

		// Find the "Close chat" toggle-btn that follows the session-picker
		const closeChatPos = src.indexOf('aria-label="Close chat"');

		expect(pickerStart).toBeGreaterThan(-1);
		expect(dropdownPos).toBeGreaterThan(-1);
		expect(closeChatPos).toBeGreaterThan(-1);

		// Dropdown must appear after session-picker opens but before the close chat button
		expect(dropdownPos).toBeGreaterThan(pickerStart);
		expect(dropdownPos).toBeLessThan(closeChatPos);
	});

	it('session-picker has position: relative in its CSS', () => {
		// Extract the CSS rule for .session-picker
		const styleBlock = fullSrc.match(/<style[\s\S]*<\/style>/);
		expect(styleBlock).not.toBeNull();

		const sessionPickerRule = styleBlock![0].match(
			/\.session-picker\s*\{[^}]*\}/
		);
		expect(sessionPickerRule).not.toBeNull();
		expect(sessionPickerRule![0]).toContain('position: relative');
	});
});
