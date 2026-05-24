interface LoadingFallbackProps {
  /** Spinner size — maps to DaisyUI loading size classes */
  size?: 'xs' | 'sm' | 'md' | 'lg';
  /** Optional loading text for screen readers */
  text?: string;
  /** Additional classes (e.g. text-primary, mb-2) */
  className?: string;
}

export function LoadingFallback({ size = 'lg', text = 'Loading', className = '' }: LoadingFallbackProps) {
  return (
    <div className="flex min-h-[60vh] items-center justify-center" role="status" aria-label={text}>
      <span className={`loading loading-spinner loading-${size} text-primary ${className}`} aria-hidden="true" />
    </div>
  );
}

/**
 * Inline loading spinner for use inside buttons, cards, or other components.
 * Not centered — renders just the spinner element.
 */
export function LoadingSpinner({ size = 'md', className = '' }: LoadingFallbackProps) {
  return <span className={`loading loading-spinner loading-${size} ${className}`} aria-hidden="true" />;
}
