/**
 * Theme store for TelePromptr.
 *
 * Uses Svelte 5 runes for reactive state management.
 * Reads initial preference from localStorage (key: `telepromptr-theme`),
 * falls back to `prefers-color-scheme` media query, and applies/removes
 * the `dark` class on `document.documentElement`.
 *
 * @module
 */

const STORAGE_KEY = 'telepromptr-theme';

/** The three possible theme modes. */
export type ThemeMode = 'light' | 'dark' | 'system';

/**
 * Determine the effective display theme based on the selected mode.
 */
function resolveTheme(mode: ThemeMode): 'light' | 'dark' {
	if (mode === 'system') {
		if (typeof window !== 'undefined') {
			return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
		}
		return 'light';
	}
	return mode;
}

/**
 * Read stored preference from localStorage.
 */
function readStoredMode(): ThemeMode {
	if (typeof window === 'undefined') return 'system';
	try {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored === 'light' || stored === 'dark' || stored === 'system') {
			return stored;
		}
	} catch {
		// localStorage may be unavailable
	}
	return 'system';
}

/**
 * Apply the resolved theme to the document root element.
 */
function applyTheme(resolved: 'light' | 'dark'): void {
	if (typeof document === 'undefined') return;
	document.documentElement.classList.toggle('dark', resolved === 'dark');
}

function createThemeState() {
	let mode = $state<ThemeMode>(readStoredMode());
	let resolved = $derived<'light' | 'dark'>(resolveTheme(mode));

	$effect(() => {
		applyTheme(resolved);
	});

	// Listen for OS theme changes when mode is 'system'
	$effect(() => {
		if (mode !== 'system' || typeof window === 'undefined') return;

		const mql = window.matchMedia('(prefers-color-scheme: dark)');
		const handler = () => {
			applyTheme(resolveTheme('system'));
		};
		mql.addEventListener('change', handler);
		return () => mql.removeEventListener('change', handler);
	});

	return {
		/** The user-selected mode: 'light', 'dark', or 'system'. */
		get mode(): ThemeMode {
			return mode;
		},

		/** Set the theme mode and persist to localStorage. */
		set mode(value: ThemeMode) {
			mode = value;
			try {
				localStorage.setItem(STORAGE_KEY, value);
			} catch {
				// localStorage may be unavailable
			}
		},

		/** The effective resolved theme: 'light' or 'dark'. */
		get resolved(): 'light' | 'dark' {
			return resolved;
		},

		/** Toggle between light and dark (ignoring system). */
		toggle(): void {
			this.mode = resolved === 'dark' ? 'light' : 'dark';
		}
	};
}

/** Singleton theme state instance. */
export const theme = createThemeState();
