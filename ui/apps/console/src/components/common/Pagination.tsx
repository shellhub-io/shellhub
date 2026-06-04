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
  if (totalPages <= 1) return null;

  return (
    <div className="flex items-center justify-between mt-4 px-1">
      <span className="text-2xs font-mono text-text-muted">
        {totalCount !== undefined
          ? `${totalCount} ${itemLabel}${totalCount !== 1 ? "s" : ""}`
          : `Page ${page} of ${totalPages}`}
      </span>
      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(page - 1)}
          disabled={page <= 1}
          className="px-2.5 py-1 text-xs font-mono text-text-secondary hover:text-text-primary disabled:opacity-soft disabled:cursor-not-allowed transition-colors"
        >
          Prev
        </button>
        <span className="text-xs font-mono text-text-muted tabular-nums px-2">
          {page} / {totalPages}
        </span>
        <button
          onClick={() => onPageChange(page + 1)}
          disabled={page >= totalPages}
          className="px-2.5 py-1 text-xs font-mono text-text-secondary hover:text-text-primary disabled:opacity-soft disabled:cursor-not-allowed transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  );
}
