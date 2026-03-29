let fullscreen = $state(false);

export function isFullscreen(): boolean {
	return fullscreen;
}

export function toggleFullscreen(): void {
	if (fullscreen) {
		document.exitFullscreen().catch(() => {});
	} else {
		document.documentElement.requestFullscreen().catch(() => {});
	}
}

export function initFullscreen(): void {
	fullscreen = !!document.fullscreenElement;
	document.addEventListener('fullscreenchange', () => {
		fullscreen = !!document.fullscreenElement;
	});
}
