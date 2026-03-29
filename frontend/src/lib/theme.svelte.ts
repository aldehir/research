export type Theme = 'light' | 'dark';

const VALID_THEMES: Theme[] = ['light', 'dark'];
const STORAGE_KEY = 'theme';

let theme = $state<Theme>('light');

export function getTheme(): Theme {
	return theme;
}

export function setTheme(t: Theme): void {
	theme = t;
	localStorage.setItem(STORAGE_KEY, t);
	applyTheme(t);
}

export function getResolvedTheme(): 'light' | 'dark' {
	return theme;
}

export function initTheme(): void {
	const stored = localStorage.getItem(STORAGE_KEY);
	if (stored && VALID_THEMES.includes(stored as Theme)) {
		theme = stored as Theme;
	} else {
		theme = 'light';
	}
	applyTheme(theme);
}

function applyTheme(t: Theme): void {
	document.documentElement.setAttribute('data-theme', t);
}
