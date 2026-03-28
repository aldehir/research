import { listPapers, uploadPaper, deletePaper } from '$lib/api';
import type { Paper } from '$lib/api';

class PapersStore {
	papers = $state<Paper[]>([]);
	selectedId = $state<string | null>(null);

	get selectedPaper(): Paper | null {
		return this.papers.find(p => p.id === this.selectedId) ?? null;
	}

	async load(): Promise<void> {
		this.papers = await listPapers();
	}

	select(id: string): void {
		this.selectedId = id;
	}

	async upload(file: File): Promise<void> {
		await uploadPaper(file);
		await this.load();
	}

	async remove(id: string): Promise<void> {
		await deletePaper(id);
		if (this.selectedId === id) {
			this.selectedId = null;
		}
		await this.load();
	}
}

export const papersStore = new PapersStore();
