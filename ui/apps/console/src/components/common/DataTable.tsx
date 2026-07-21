import { Fragment, type ReactNode } from "react";
import {
  ChevronUpIcon,
  ChevronDownIcon,
  ChevronUpDownIcon,
} from "@heroicons/react/24/outline";
import { Card } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import Pagination from "./Pagination";
import PageLoader from "@/components/common/PageLoader";

const TH_CLASS =
  "px-4 py-3 text-left text-2xs font-mono font-semibold uppercase tracking-compact text-text-muted whitespace-nowrap";

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

  /**
   * Groups rows into labelled sections: a full-width header row is inserted before the first row of
   * each section (i.e. whenever `sectionOf` changes from the previous row). `data` must already be
   * ordered so each section's rows are contiguous. `sectionLabel` maps a section key to its heading.
   */
  sectionOf?: (row: T) => string;
  sectionLabel?: (section: string) => string;

  /**
   * A full-width block appended after the last data row. For a grouped table whose trailing section
   * is empty — e.g. user-created rows sitting under always-present built-in rows — pass an onboarding
   * placeholder here; it stands in for the whole section (no section header, since the placeholder
   * already names it). The caller gates this on the true, unpaginated count (page data alone can't
   * tell "empty" from "on another page").
   */
  trailingEmptyState?: ReactNode;

  /** When a row's key matches, an extra full-width row is rendered below it with
   * `renderExpandedRow`. Used for inline accordion panels. */
  expandedRowKey?: string | null;
  renderExpandedRow?: (row: T) => ReactNode;

  /** Hide the column header row. For compact/embedded previews where the columns
   * are self-evident and a header would read as a separate table. */
  hideHeader?: boolean;

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
  /**
   * When true, the header row gets a top rule so it reads as a boundary in an
   * embedded context with no surrounding border. Keep it off (the default) when
   * the parent already draws a border, or the two rules stack into a thick edge.
   */
  headerTopBorder?: boolean;
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
  sectionOf,
  sectionLabel,
  trailingEmptyState,
  expandedRowKey,
  renderExpandedRow,
  hideHeader = false,
  isLoading,
  loadingMessage = "Loading...",
  emptyState,
  emptyMessage = "No data available",
  noWrapper = false,
  headerTopBorder = false,
}: DataTableProps<T>) {
  const hasPagination =
    page !== undefined &&
    totalPages !== undefined &&
    onPageChange !== undefined;

  const tableContent = (
    <div className="overflow-x-auto">
      <table className="w-full" aria-label={label}>
        {!hideHeader && (
          <thead>
            <tr
              className={cn(
                "border-b border-border bg-surface/50",
                // Only an embedded table with no surrounding border asks for a top rule; keying it off
                // noWrapper would double the border on callers embedded inside an already-bordered card.
                headerTopBorder && "border-t",
              )}
            >
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
        )}
        <tbody className="divide-y divide-border/60">
          {isLoading && data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="px-4 py-16 text-center">
                <PageLoader label={loadingMessage} showLabel padding="none" />
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
            <>
              {data.map((row, index) => {
                const extraClass = rowClassName?.(row) ?? "";
                const clickHandler = onRowClick;
                const isClickable = !!clickHandler;
                const key = rowKey(row, index);
                const isExpanded =
                  !!renderExpandedRow && expandedRowKey === key;
                // Insert a section header before the first row of each section (when it changes from the
                // previous row). `data` is assumed grouped so each section is contiguous.
                const section = sectionOf?.(row);
                const showSection =
                  section !== undefined &&
                  section !==
                    (index > 0 ? sectionOf?.(data[index - 1]) : undefined);
                return (
                  <Fragment key={key}>
                    {showSection && (
                      <tr>
                        <td
                          colSpan={columns.length}
                          className="px-4 pt-2 pb-2 font-mono text-2xs uppercase tracking-label text-text-muted/70"
                        >
                          {sectionLabel?.(section) ?? section}
                        </td>
                      </tr>
                    )}
                    <tr
                      onClick={
                        clickHandler ? () => clickHandler(row) : undefined
                      }
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
                    {isExpanded && (
                      <tr>
                        <td colSpan={columns.length} className="p-0">
                          {renderExpandedRow(row)}
                        </td>
                      </tr>
                    )}
                  </Fragment>
                );
              })}
              {trailingEmptyState && (
                <tr>
                  <td colSpan={columns.length} className="p-0">
                    {trailingEmptyState}
                  </td>
                </tr>
              )}
            </>
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
        <Card className="overflow-hidden">{tableContent}</Card>
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
