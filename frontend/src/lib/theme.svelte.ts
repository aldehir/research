export type Theme = 'light' | 'dark' | 'system';

const VALID_THEMES: Theme[] = ['light', 'dark', 'system'];
const STORAGE_KEY = 'theme';

let theme = $state<Theme>('system');

export function getTheme(): Theme {
	return theme;
}

export function setTheme(t: Theme): void {
	theme = t;
	localStorage.setItem(STORAGE_KEY, t);
	applyTheme(t);
}

export function getResolvedTheme(): 'light' | 'dark' {
	if (theme === 'system') {
		return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
	}
	return theme;
}

export function initTheme(): void {
	const stored = localStorage.getItem(STORAGE_KEY);
	if (stored && VALID_THEMES.includes(stored as Theme)) {
		theme = stored as Theme;
	} else {
		theme = 'system';
	}
	applyTheme(theme);
}

function applyTheme(t: Theme): void {
	if (t === 'system') {
		document.documentElement.removeAttribute('data-theme');
	} else {
		document.documentElement.setAttribute('data-theme', t);
	}
}
