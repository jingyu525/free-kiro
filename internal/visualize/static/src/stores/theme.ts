// stores/theme.ts — current theme value (dark/light/auto).
// Drives <html data-theme="..."> via lib/theme.ts.

import { Signal } from '../lib/signal';

export type Theme = 'dark' | 'light' | 'auto';

export const themeStore = new Signal<Theme>('auto');
