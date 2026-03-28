// @vitest-environment jsdom
import { describe, it, expect } from 'vitest';
import { renderMarkdown } from '$lib/markdown';

describe('renderMarkdown', () => {
	describe('basic formatting', () => {
		it('renders headings', () => {
			const html = renderMarkdown('# Hello');
			expect(html).toContain('<h1');
			expect(html).toContain('Hello');
		});

		it('renders h2 headings', () => {
			const html = renderMarkdown('## Subheading');
			expect(html).toContain('<h2');
			expect(html).toContain('Subheading');
		});

		it('renders bold text', () => {
			const html = renderMarkdown('This is **bold** text');
			expect(html).toContain('<strong>bold</strong>');
		});

		it('renders italic text', () => {
			const html = renderMarkdown('This is *italic* text');
			expect(html).toContain('<em>italic</em>');
		});

		it('renders inline code', () => {
			const html = renderMarkdown('Use `console.log()` here');
			expect(html).toContain('<code>console.log()</code>');
		});

		it('renders links', () => {
			const html = renderMarkdown('[click here](https://example.com)');
			expect(html).toContain('<a');
			expect(html).toContain('href="https://example.com"');
			expect(html).toContain('click here');
		});

		it('renders paragraphs', () => {
			const html = renderMarkdown('First paragraph\n\nSecond paragraph');
			expect(html).toContain('<p>First paragraph</p>');
			expect(html).toContain('<p>Second paragraph</p>');
		});
	});

	describe('code blocks', () => {
		it('renders fenced code blocks', () => {
			const html = renderMarkdown('```\nconst x = 1;\n```');
			expect(html).toContain('<pre>');
			expect(html).toContain('<code>');
			expect(html).toContain('const x = 1;');
		});

		it('renders code blocks with language annotation', () => {
			const html = renderMarkdown('```javascript\nconst x = 1;\n```');
			expect(html).toContain('<code');
			expect(html).toContain('javascript');
			expect(html).toContain('const x = 1;');
		});
	});

	describe('lists', () => {
		it('renders unordered lists', () => {
			const html = renderMarkdown('- item one\n- item two\n- item three');
			expect(html).toContain('<ul>');
			expect(html).toContain('<li>');
			expect(html).toContain('item one');
			expect(html).toContain('item two');
			expect(html).toContain('item three');
		});

		it('renders ordered lists', () => {
			const html = renderMarkdown('1. first\n2. second\n3. third');
			expect(html).toContain('<ol>');
			expect(html).toContain('<li>');
			expect(html).toContain('first');
			expect(html).toContain('second');
		});

		it('renders nested lists', () => {
			const html = renderMarkdown('- parent\n  - child\n  - child2');
			expect(html).toContain('<ul>');
			expect(html).toContain('parent');
			expect(html).toContain('child');
		});
	});

	describe('other elements', () => {
		it('renders blockquotes', () => {
			const html = renderMarkdown('> This is a quote');
			expect(html).toContain('<blockquote>');
			expect(html).toContain('This is a quote');
		});

		it('renders horizontal rules', () => {
			const html = renderMarkdown('---');
			expect(html).toContain('<hr');
		});

		it('renders tables', () => {
			const md = '| Header | Value |\n| --- | --- |\n| row1 | data1 |';
			const html = renderMarkdown(md);
			expect(html).toContain('<table>');
			expect(html).toContain('<th>');
			expect(html).toContain('Header');
			expect(html).toContain('data1');
		});
	});

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

	describe('streaming resilience', () => {
		it('handles incomplete code fences gracefully', () => {
			const html = renderMarkdown('Here is code:\n```python\ndef hello():');
			// Should not throw and should render something reasonable
			expect(html).toBeDefined();
			expect(html).toContain('def hello():');
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
