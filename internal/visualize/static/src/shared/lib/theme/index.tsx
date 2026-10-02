// shared/lib/theme — theme Context + persistence + system preference.
//
// Lives in shared (not app) because theme is browser-environment plumbing,
// not business semantics. features/theme-toggle may consume it without
// breaking FSD's top-down import direction.

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

export type ThemeValue = 'dark' | 'light' | 'auto';

const STORAGE_KEY = 'fk-theme';

function systemPrefersDark(): boolean {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches;
}

function resolveTheme(value: ThemeValue): 'dark' | 'light' {
  if (value === 'auto') return systemPrefersDark() ? 'dark' : 'light';
  return value;
}

function applyToHtml(resolved: 'dark' | 'light'): void {
  document.documentElement.setAttribute('data-theme', resolved);
}

interface ThemeApi {
  theme: ThemeValue;
  resolved: 'dark' | 'light';
  setTheme: (next: ThemeValue) => void;
}

const ThemeContext = createContext<ThemeApi | null>(null);

export function ThemeProvider({ children }: { children: ReactNode }): JSX.Element {
  const [theme, setThemeState] = useState<ThemeValue>(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === 'dark' || stored === 'light' || stored === 'auto') return stored;
    return 'auto';
  });
  const [resolved, setResolved] = useState<'dark' | 'light'>(() => resolveTheme(theme));

  useEffect(() => {
    const r = resolveTheme(theme);
    setResolved(r);
    applyToHtml(r);
    localStorage.setItem(STORAGE_KEY, theme);
  }, [theme]);

  useEffect(() => {
    if (theme !== 'auto') return;
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const onChange = (): void => {
      const r = resolveTheme('auto');
      setResolved(r);
      applyToHtml(r);
    };
    mq.addEventListener('change', onChange);
    return () => mq.removeEventListener('change', onChange);
  }, [theme]);

  const setTheme = useCallback((next: ThemeValue): void => setThemeState(next), []);

  const value = useMemo<ThemeApi>(() => ({ theme, resolved, setTheme }), [theme, resolved, setTheme]);
  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeApi {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error('useTheme must be used inside <ThemeProvider>');
  return ctx;
}