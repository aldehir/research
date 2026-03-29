const STORAGE_KEY = 'fullscreen';

let fullscreen = $state(false);

export function isFullscreen(): boolean {
	return fullscreen;
}

export function toggleFullscreen(): void {
	setFullscreen(!fullscreen);
}

export function setFullscreen(value: boolean): void {
	fullscreen = value;
	localStorage.setItem(STORAGE_KEY, String(value));
	applyFullscreen(value);
}

export function initFullscreen(): void {
	const stored = localStorage.getItem(STORAGE_KEY);
	fullscreen = stored === 'true';
	applyFullscreen(fullscreen);
}

function applyFullscreen(value: boolean): void {
	if (value) {
		document.documentElement.setAttribute('data-fullscreen', '');
	} else {
		document.documentElement.removeAttribute('data-fullscreen');
	}
}
