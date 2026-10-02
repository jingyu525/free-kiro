// features/theme-toggle — light/dark/auto cycle button.

import { useTheme } from '@shared/lib/theme';
import { Button } from '@shared/ui/Button';

const ORDER = ['dark', 'light', 'auto'] as const;
type ThemeValue = (typeof ORDER)[number];

const ICON: Record<ThemeValue, string> = {
  dark: '🌙',
  light: '☀️',
  auto: '🖥️',
};

const LABEL: Record<ThemeValue, string> = {
  dark: 'dark theme',
  light: 'light theme',
  auto: 'auto theme',
};

export function ThemeToggle(): JSX.Element {
  const { theme, setTheme } = useTheme();
  const cycle = (): void => {
    const idx = ORDER.indexOf(theme);
    const next = ORDER[(idx + 1) % ORDER.length] ?? 'dark';
    setTheme(next);
  };
  return (
    <Button
      variant="ghost"
      onClick={cycle}
      aria-label={`Switch theme (current: ${LABEL[theme]})`}
      title={LABEL[theme]}
    >
      <span aria-hidden="true">{ICON[theme]}</span>
    </Button>
  );
}