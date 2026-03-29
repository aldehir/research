import { marked } from 'marked';
import DOMPurify from 'dompurify';
import hljs from 'highlight.js';

marked.setOptions({
	gfm: true,
	breaks: false,
});

const renderer = new marked.Renderer();
renderer.code = ({ text, lang }: { text: string; lang?: string }) => {
	if (lang && hljs.getLanguage(lang)) {
		const highlighted = hljs.highlight(text, { language: lang }).value;
		return `<pre><code class="hljs language-${lang}">${highlighted}</code></pre>`;
	}
	const escaped = text
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;');
	return `<pre><code>${escaped}</code></pre>`;
};

const purify = DOMPurify(window);
purify.addHook('uponSanitizeAttribute', (node, data) => {
	if (node.tagName === 'SPAN' && data.attrName === 'class') {
		const classes = data.attrValue.split(/\s+/);
		if (classes.every(c => c.startsWith('hljs-') || c === 'hljs')) {
			data.forceKeepAttr = true;
		}
	}
	if (node.tagName === 'CODE' && data.attrName === 'class') {
		const classes = data.attrValue.split(/\s+/);
		if (classes.every(c => c.startsWith('hljs') || c.startsWith('language-'))) {
			data.forceKeepAttr = true;
		}
	}
});

export function renderMarkdown(content: string): string {
	if (!content) return '';
	const raw = marked.parse(content, { async: false, renderer }) as string;
	return purify.sanitize(raw);
}
