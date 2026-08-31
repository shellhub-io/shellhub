import { useState } from "react";

/**
 * Which way a column sorts.
 */
export type SortOrder = "asc" | "desc";

interface UseTableSortOptions<TField extends string> {
  defaultField: TField;
  defaultOrder?: SortOrder;
  onSortChange?: () => void;
}

/**
 * Tracks a table's sort column and direction. onSortChange fires on a change so the caller can
 * return to the first page — a sorted list re-paginated from where it was shows the wrong rows.
 */
export function useTableSort<TField extends string>({
  defaultField,
  defaultOrder = "desc",
  onSortChange,
}: UseTableSortOptions<TField>) {
  const [sortBy, setSortBy] = useState<TField>(defaultField);
  const [orderBy, setOrderBy] = useState<SortOrder>(defaultOrder);

  const handleSort = (field: string) => {
    const f = field as TField;
    if (sortBy === f) {
      setOrderBy((prev) => (prev === "asc" ? "desc" : "asc"));
    } else {
      setSortBy(f);
      setOrderBy(f === "name" ? "asc" : "desc");
    }
    onSortChange?.();
  };

  return { sortBy, orderBy, handleSort };
}
