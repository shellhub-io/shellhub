import { describe, it, expect } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useEditModal } from "../useEditModal";

describe("useEditModal", () => {
  it("opens on the row being edited", () => {
    const { result } = renderHook(() => useEditModal<{ id: string }>());

    act(() => result.current.openEdit({ id: "row-1" }));

    expect(result.current.isOpen).toBe(true);
    expect(result.current.target).toEqual({ id: "row-1" });
  });

  it("opens empty for a new row after editing one", () => {
    const { result } = renderHook(() => useEditModal<{ id: string }>());

    act(() => result.current.openEdit({ id: "row-1" }));
    act(() => result.current.close());
    act(() => result.current.openNew());

    expect(result.current.isOpen).toBe(true);
    expect(result.current.target).toBeNull();
  });
});
