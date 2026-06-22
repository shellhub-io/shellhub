interface Props {
  page: number;
  totalPages: number;
  totalCount?: number;
  itemLabel?: string;
  onPageChange: (page: number) => void;
}

export default function Pagination({
  page,
  totalPages,
  totalCount,
  itemLabel = "item",
  onPageChange,
}: Props) {
  // Nothing to render: no navigation needed and no positive count to show.
  // (Empty lists — totalCount 0 — defer to each page's own empty-state.)
  if (totalPages <= 1 && !totalCount) return null;

  const countLabel =
    totalCount !== undefined
      ? `${totalCount} ${itemLabel}${totalCount !== 1 ? "s" : ""}`
      : null;

  const leftLabel =
    countLabel ??
    (totalPages > 1 ? `Page ${page} of ${totalPages}` : null);

  return (
    <nav
      aria-label="Pagination"
      className="flex items-center justify-between mt-4 px-1"
    >
      {leftLabel ? (
        <span className="text-xs font-mono text-text-muted">{leftLabel}</span>
      ) : (
        <span />
      )}
      {totalPages > 1 && (
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => onPageChange(page - 1)}
            disabled={page <= 1}
            aria-label="Previous page"
            className="px-2.5 py-1 text-xs font-medium text-text-secondary hover:text-text-primary disabled:opacity-soft disabled:cursor-not-allowed transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary/50 rounded"
          >
            Prev
          </button>
          <span
            className="text-xs font-mono text-text-muted tabular-nums px-2"
            aria-current="page"
          >
            {page} / {totalPages}
          </span>
          <button
            type="button"
            onClick={() => onPageChange(page + 1)}
            disabled={page >= totalPages}
            aria-label="Next page"
            className="px-2.5 py-1 text-xs font-medium text-text-secondary hover:text-text-primary disabled:opacity-soft disabled:cursor-not-allowed transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary/50 rounded"
          >
            Next
          </button>
        </div>
      )}
    </nav>
  );
}
