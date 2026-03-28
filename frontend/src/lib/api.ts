export interface Paper {
	id: string;
	title: string;
	file_path: string;
	file_size: number;
	created_at: string;
}

async function handleResponse<T>(response: Response): Promise<T> {
	if (!response.ok) {
		const body = await response.json() as { error: string };
		throw new Error(body.error);
	}
	return response.json() as Promise<T>;
}

export async function listPapers(): Promise<Paper[]> {
	const response = await fetch('/api/papers');
	return handleResponse<Paper[]>(response);
}

export async function uploadPaper(file: File): Promise<Paper> {
	const formData = new FormData();
	formData.append('file', file);
	const response = await fetch('/api/papers', {
		method: 'POST',
		body: formData
	});
	return handleResponse<Paper>(response);
}

export async function getPaper(id: string): Promise<Paper> {
	const response = await fetch(`/api/papers/${id}`);
	return handleResponse<Paper>(response);
}

export async function deletePaper(id: string): Promise<void> {
	const response = await fetch(`/api/papers/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		const body = await response.json() as { error: string };
		throw new Error(body.error);
	}
}

export function getPdfUrl(id: string): string {
	return `/api/papers/${id}/pdf`;
}
