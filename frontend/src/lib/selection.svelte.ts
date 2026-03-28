let selectedText = $state('');
let surroundingText = $state('');

export function setSelection(selected: string, surrounding: string): void {
	selectedText = selected;
	surroundingText = surrounding;
}

export function clearSelection(): void {
	selectedText = '';
	surroundingText = '';
}

export function getSelectedText(): string {
	return selectedText;
}

export function getSurroundingText(): string {
	return surroundingText;
}
