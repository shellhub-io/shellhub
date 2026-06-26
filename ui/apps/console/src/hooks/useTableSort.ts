import { useState } from "react";

export type SortOrder = "asc" | "desc";

interface UseTableSortOptions<TField extends string> {
  /** Field selected on first render. */
  defaultField: TField;
  /** Order selected on first render. Defaults to "desc". */
  defaultOrder?: SortOrder;
  /** Side effect to run after any sort change (e.g. reset pagination to page 1). */
  onSortChange?: () => void;
}

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
