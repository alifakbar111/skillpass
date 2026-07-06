interface SkeletonProps {
  /** Number of skeleton rows/lines to render */
  count?: number;
  /** Height of each skeleton line */
  height?: string;
  /** Width — defaults to full */
  width?: string;
  /** Additional classes */
  className?: string;
}

function Line({ height = 'h-4', width = 'w-full', className = '' }: SkeletonProps) {
  return <div className={`skeleton ${height} ${width} ${className}`} />;
}

/**
 * Table skeleton — mimics a table with header + rows.
 */
export function TableSkeleton({ rows = 5, cols = 6 }: { rows?: number; cols?: number }) {
  return (
    <div className="space-y-3">
      {/* Header */}
      <div className="flex gap-4">
        {Array.from({ length: cols }).map((_, i) => (
          <Line key={`h-${i}`} height="h-4" width={`${Math.max(60, 100 / cols)}%`} className="opacity-60" />
        ))}
      </div>
      <div className="divider my-0" />
      {/* Rows */}
      {Array.from({ length: rows }).map((_, r) => (
        <div key={`r-${r}`} className="flex gap-4">
          {Array.from({ length: cols }).map((_, c) => (
            <Line key={`c-${r}-${c}`} height="h-3" width={`${Math.max(40, 100 / cols)}%`} className="opacity-30" />
          ))}
        </div>
      ))}
    </div>
  );
}

/**
 * Card grid skeleton — mimics stat cards or grid layout.
 */
export function CardGridSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="card bg-base-100 border border-base-300 p-6 space-y-3">
          <Line height="h-3" width="w-1/3" className="opacity-40" />
          <Line height="h-8" width="w-1/2" />
          <Line height="h-3" width="w-2/3" className="opacity-30" />
          <div className="skeleton h-2 w-full rounded-full opacity-20" />
        </div>
      ))}
    </div>
  );
}

/**
 * Detail form skeleton — mimics a two-column form layout.
 */
export function DetailSkeleton() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="card bg-base-200 p-6 space-y-3">
          <Line height="h-5" width="w-1/3" className="opacity-50" />
          {Array.from({ length: 4 }).map((_, j) => (
            <div key={j} className="space-y-1">
              <Line height="h-2" width="w-1/4" className="opacity-30" />
              <Line height="h-8" width="w-full" className="opacity-20" />
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}

/**
 * Profile header skeleton.
 */
export function HeaderSkeleton() {
  return (
    <div className="flex items-center gap-4 mb-6">
      <div className="skeleton h-10 w-10 rounded-full opacity-30" />
      <div className="space-y-2 flex-1">
        <Line height="h-6" width="w-1/3" />
        <Line height="h-3" width="w-1/2" className="opacity-40" />
      </div>
    </div>
  );
}
