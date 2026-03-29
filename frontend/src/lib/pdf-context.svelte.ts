let currentPage = $state(1);

export function setCurrentPage(page: number): void {
	currentPage = page;
}

export function getCurrentPage(): number {
	return currentPage;
}
