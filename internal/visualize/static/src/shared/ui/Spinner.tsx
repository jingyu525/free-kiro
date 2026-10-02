// shared/ui/Spinner — pure-CSS spinner; respects prefers-reduced-motion.

interface SizeProps {
  size?: 'sm' | 'md' | 'lg';
  label?: string;
}

export function Spinner({ size = 'md', label = 'Loading' }: SizeProps): JSX.Element {
  return (
    <span
      className={`spinner spinner-${size}`}
      role="status"
      aria-label={label}
      aria-live="polite"
    />
  );
}