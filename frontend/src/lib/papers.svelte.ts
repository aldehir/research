import { listPapers, uploadPaper, deletePaper } from '$lib/api';
import type { Paper } from '$lib/api';

class PapersStore {
	papers = $state<Paper[]>([]);
	loading = $state(false);
	selectedId = $state<string | null>(null);

	get selectedPaper(): Paper | null {
		return this.papers.find(p => p.id === this.selectedId) ?? null;
	}

	async load(): Promise<void> {
		this.loading = true;
		try {
			this.papers = await listPapers();
		} finally {
			this.loading = false;
		}
	}

	select(id: string): void {
		this.selectedId = id;
	}

	deselect(): void {
		this.selectedId = null;
	}

	async loadAndSelect(id: string): Promise<void> {
		if (this.papers.length === 0) {
			await this.load();
		}
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
