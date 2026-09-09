import { describe, it, expect, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useTableSort } from "@/hooks/useTableSort";

function renderTableSort(onSortChange?: () => void) {
  return renderHook(() =>
    useTableSort<"name" | "last_seen">({
      defaultField: "last_seen",
      onSortChange,
    }),
  );
}

describe("useTableSort", () => {
  it("starts on the default field, descending", () => {
    const { result } = renderTableSort();

    expect(result.current.sortBy).toBe("last_seen");
    expect(result.current.orderBy).toBe("desc");
  });

  it("flips the direction when the sorted column is chosen again", () => {
    const { result } = renderTableSort();

    act(() => result.current.handleSort("last_seen"));
    expect(result.current.orderBy).toBe("asc");

    act(() => result.current.handleSort("last_seen"));
    expect(result.current.orderBy).toBe("desc");
  });

  it("switching column restarts the direction — name ascending, anything else descending", () => {
    const { result } = renderTableSort();

    act(() => result.current.handleSort("name"));
    expect(result.current.sortBy).toBe("name");
    expect(result.current.orderBy).toBe("asc");

    act(() => result.current.handleSort("last_seen"));
    expect(result.current.sortBy).toBe("last_seen");
    expect(result.current.orderBy).toBe("desc");
  });

  it("reports every sort change so the caller can go back to page one", () => {
    const onSortChange = vi.fn();
    const { result } = renderTableSort(onSortChange);

    act(() => result.current.handleSort("name"));
    act(() => result.current.handleSort("name"));

    expect(onSortChange).toHaveBeenCalledTimes(2);
  });
});
