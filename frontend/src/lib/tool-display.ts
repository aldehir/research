const toolLabels: Record<string, string> = {
	search_pdf: 'Searched PDF',
	read_page: 'Read page',
	go_to_page: 'Navigated to page',
	snapshot_page: 'Page snapshot',
};

export function formatToolLabel(name: string): string {
	return toolLabels[name] ?? 'Used tool';
}

export function formatToolArgs(name: string, args: Record<string, unknown>): string {
	switch (name) {
		case 'search_pdf':
			return `"${args.query}"`;

		case 'read_page':
		case 'go_to_page':
		case 'snapshot_page':
			return `page ${args.page}`;
		default:
			return JSON.stringify(args);
	}
}
