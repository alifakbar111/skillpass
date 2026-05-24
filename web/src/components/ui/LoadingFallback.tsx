export function LoadingFallback() {
  return (
    <div className="flex min-h-[60vh] items-center justify-center" role="status" aria-label="Loading content">
      <span className="loading loading-spinner loading-lg text-primary" aria-hidden="true" />
    </div>
  );
}
