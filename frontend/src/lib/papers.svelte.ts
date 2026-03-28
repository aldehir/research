import { listPapers, uploadPaper, deletePaper } from '$lib/api';
import type { Paper } from '$lib/api';

let papers = $state<Paper[]>([]);
let selectedPaperId = $state<string | null>(null);

const selectedPaper = $derived(papers.find(p => p.id === selectedPaperId) ?? null);

export async function loadPapers(): Promise<void> {
	papers = await listPapers();
}

export function selectPaper(id: string): void {
	selectedPaperId = id;
}

export async function upload(file: File): Promise<void> {
	await uploadPaper(file);
	await loadPapers();
}

export async function remove(id: string): Promise<void> {
	await deletePaper(id);
	if (selectedPaperId === id) {
		selectedPaperId = null;
	}
	await loadPapers();
}

export function getPapers(): Paper[] {
	return papers;
}

export function getSelectedPaper(): Paper | null {
	return selectedPaper;
}
