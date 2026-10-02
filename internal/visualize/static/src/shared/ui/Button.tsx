// shared/ui/Button — generic button used by features (refresh, retry, theme toggle).

import type { ButtonHTMLAttributes, ReactNode } from 'react';

export type ButtonVariant = 'primary' | 'ghost' | 'danger';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  spinning?: boolean;
  children: ReactNode;
}

export function Button({
  variant = 'ghost',
  spinning = false,
  disabled,
  children,
  className,
  ...rest
}: ButtonProps): JSX.Element {
  const classes = ['btn', `btn-${variant}`, spinning ? 'spinning' : '', className ?? '']
    .filter(Boolean)
    .join(' ');
  return (
    <button
      type="button"
      className={classes}
      disabled={disabled || spinning}
      aria-busy={spinning || undefined}
      {...rest}
    >
      {children}
    </button>
  );
}