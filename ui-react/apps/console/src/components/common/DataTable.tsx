import type { ReactNode } from "react";
import {
  ChevronUpIcon,
  ChevronDownIcon,
  ChevronUpDownIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@/utils/cn";
import Pagination from "./Pagination";

const TH_CLASS =
  "px-4 py-3 text-left text-2xs font-mono font-semibold uppercase tracking-compact text-text-muted whitespace-nowrap";

const DEFAULT_WRAPPER =
  "bg-card border border-border rounded-xl overflow-hidden";

export interface Column<T> {
  key: string;
  header: string;
  headerClassName?: string;
  sortable?: boolean;
  render: (row: T) => ReactNode;
}

export interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  rowKey: (row: T, index: number) => string;

  label?: string;

  page?: number;
  totalPages?: number;
  totalCount?: number;
  itemLabel?: string;
  onPageChange?: (page: number) => void;

  sortField?: string;
  sortOrder?: "asc" | "desc";
  onSort?: (field: string) => void;

  onRowClick?: (row: T) => void;
  rowClassName?: (row: T) => string | undefined;

  isLoading?: boolean;
  loadingMessage?: string;
  emptyState?: ReactNode;
  emptyMessage?: string;

  /**
   * When true, the default card wrapper (background, border, rounded corners)
   * is not rendered. Use for embedded contexts where the parent already
   * provides the chrome.
   */
  noWrapper?: boolean;
}

function SortIndicator({
  field,
  sortField,
  sortOrder,
}: {
  field: string;
  sortField?: string;
  sortOrder?: "asc" | "desc";
}) {
  if (field !== sortField) {
    return (
      <ChevronUpDownIcon
        className="w-3 h-3 inline ml-0.5 text-text-muted/40"
        strokeWidth={2.5}
      />
    );
  }
  if (sortOrder === "asc") {
    return (
      <ChevronUpIcon className="w-3 h-3 inline ml-0.5" strokeWidth={2.5} />
    );
  }
  if (sortOrder === "desc") {
    return (
      <ChevronDownIcon className="w-3 h-3 inline ml-0.5" strokeWidth={2.5} />
    );
  }
  return null;
}

function getAriaSort(
  colKey: string,
  sortField: string | undefined,
  sortOrder: "asc" | "desc" | undefined,
): "ascending" | "descending" | "none" {
  if (colKey !== sortField) return "none";
  if (sortOrder === "asc") return "ascending";
  if (sortOrder === "desc") return "descending";
  return "none";
}

export default function DataTable<T>({
  columns,
  data,
  rowKey,
  label,
  page,
  totalPages,
  totalCount,
  itemLabel,
  onPageChange,
  sortField,
  sortOrder,
  onSort,
  onRowClick,
  rowClassName,
  isLoading,
  loadingMessage = "Loading...",
  emptyState,
  emptyMessage = "No data available",
  noWrapper = false,
}: DataTableProps<T>) {
  const hasPagination =
    page !== undefined &&
    totalPages !== undefined &&
    onPageChange !== undefined;

  const tableContent = (
    <div className="overflow-x-auto">
      <table className="w-full" aria-label={label}>
        <thead>
          <tr className="border-b border-border bg-surface/50">
            {columns.map((col) => {
              const isSortable = !!(col.sortable && onSort);
              return (
                <th
                  key={col.key}
                  className={cn(TH_CLASS, col.headerClassName)}
                  aria-sort={
                    isSortable
                      ? getAriaSort(col.key, sortField, sortOrder)
                      : undefined
                  }
                >
                  {isSortable ? (
                    <button
                      type="button"
                      onClick={() => onSort(col.key)}
                      aria-label={`Sort by ${col.header}`}
                      className="text-2xs font-mono font-semibold uppercase tracking-compact text-text-muted inline-flex items-center rounded-sm hover:text-text-primary transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary/50"
                    >
                      {col.header}
                      <SortIndicator
                        field={col.key}
                        sortField={sortField}
                        sortOrder={sortOrder}
                      />
                    </button>
                  ) : (
                    col.header
                  )}
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody className="divide-y divide-border/60">
          {isLoading && data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="px-4 py-16 text-center">
                <div
                  className="flex items-center justify-center gap-3"
                  role="status"
                  aria-live="polite"
                >
                  <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                  <span className="text-xs font-mono text-text-muted">
                    {loadingMessage}
                  </span>
                </div>
              </td>
            </tr>
          ) : data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="px-4 py-16 text-center">
                {emptyState ?? (
                  <p className="text-xs font-mono text-text-muted">
                    {emptyMessage}
                  </p>
                )}
              </td>
            </tr>
          ) : (
            data.map((row, index) => {
              const extraClass = rowClassName?.(row) ?? "";
              const clickHandler = onRowClick;
              const isClickable = !!clickHandler;
              return (
                <tr
                  key={rowKey(row, index)}
                  onClick={clickHandler ? () => clickHandler(row) : undefined}
                  onKeyDown={
                    clickHandler
                      ? (e) => {
                          if (e.key === "Enter" || e.key === " ") {
                            e.preventDefault();
                            clickHandler(row);
                          }
                        }
                      : undefined
                  }
                  tabIndex={isClickable ? 0 : undefined}
                  className={cn(
                    "group transition-colors",
                    isClickable &&
                      "cursor-pointer hover:bg-hover-subtle focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary/50 focus-visible:-outline-offset-2",
                    extraClass,
                  )}
                >
                  {columns.map((col) => (
                    <td key={col.key} className="px-4 py-3.5">
                      {col.render(row)}
                    </td>
                  ))}
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );

  return (
    <>
      {noWrapper ? (
        tableContent
      ) : (
        <div className={DEFAULT_WRAPPER}>{tableContent}</div>
      )}

      {hasPagination && (
        <Pagination
          page={page}
          totalPages={totalPages}
          totalCount={totalCount}
          itemLabel={itemLabel}
          onPageChange={onPageChange}
        />
      )}
    </>
  );
}
