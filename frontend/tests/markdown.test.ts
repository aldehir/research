// @vitest-environment jsdom
import { describe, it, expect } from 'vitest';
import { renderMarkdown } from '$lib/markdown';

describe('renderMarkdown', () => {
	describe('XSS sanitization', () => {
		it('strips script tags', () => {
			const html = renderMarkdown('<script>alert("xss")</script>');
			expect(html).not.toContain('<script>');
			expect(html).not.toContain('alert');
		});

		it('strips onerror attributes', () => {
			const html = renderMarkdown('<img src=x onerror="alert(1)">');
			expect(html).not.toContain('onerror');
		});

		it('strips javascript: URLs', () => {
			const html = renderMarkdown('[click](javascript:alert(1))');
			expect(html).not.toContain('javascript:');
		});

		it('strips event handlers in HTML', () => {
			const html = renderMarkdown('<div onclick="alert(1)">text</div>');
			expect(html).not.toContain('onclick');
		});
	});

	describe('syntax highlighting', () => {
		it('applies hljs classes to fenced code blocks', () => {
			const html = renderMarkdown('```javascript\nconst x = 1;\n```');
			expect(html).toContain('hljs');
		});

		it('preserves hljs span elements through DOMPurify', () => {
			const html = renderMarkdown('```javascript\nconst x = 1;\n```');
			expect(html).toContain('<span class="hljs-');
		});

		it('applies language class to code element', () => {
			const html = renderMarkdown('```python\ndef foo(): pass\n```');
			expect(html).toContain('language-python');
		});

		it('handles code blocks without a language specifier', () => {
			const html = renderMarkdown('```\nsome code\n```');
			expect(html).toContain('<pre>');
			expect(html).toContain('<code>');
		});
	});

	describe('streaming resilience', () => {
		it('handles incomplete code fences gracefully', () => {
			const html = renderMarkdown('Here is code:\n```python\ndef hello():');
			// Should not throw and should render something reasonable
			expect(html).toBeDefined();
			expect(html).toContain('hello');
		});

		it('handles incomplete bold syntax', () => {
			const html = renderMarkdown('This is **bold but not clos');
			expect(html).toBeDefined();
			expect(html).toContain('bold but not clos');
		});

		it('handles incomplete italic syntax', () => {
			const html = renderMarkdown('This is *italic but not clos');
			expect(html).toBeDefined();
			expect(html).toContain('italic but not clos');
		});

		it('handles incomplete link syntax', () => {
			const html = renderMarkdown('See [this link](http://');
			expect(html).toBeDefined();
		});

		it('handles empty content', () => {
			const html = renderMarkdown('');
			expect(html).toBe('');
		});

		it('handles plain text', () => {
			const html = renderMarkdown('just plain text');
			expect(html).toContain('just plain text');
		});
	});
});
