import { marked } from 'marked';
import DOMPurify from 'dompurify';

marked.setOptions({
	gfm: true,
	breaks: false,
});

const purify = DOMPurify(window);

export function renderMarkdown(content: string): string {
	if (!content) return '';
	const raw = marked.parse(content, { async: false }) as string;
	return purify.sanitize(raw);
}
