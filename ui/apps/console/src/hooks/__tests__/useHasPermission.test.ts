import { describe, it, expect, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useAuthStore } from "@/stores/authStore";
import { useHasPermission } from "../useHasPermission";

beforeEach(() => {
  useAuthStore.setState({ role: null });
});

describe("useHasPermission", () => {
  it("allows when no permission is required", () => {
    const { result } = renderHook(() => useHasPermission());
    expect(result.current).toBe(true);
  });

  it("denies when there is no role", () => {
    const { result } = renderHook(() => useHasPermission("DeviceConnect"));
    expect(result.current).toBe(false);
  });

  it("re-evaluates when the role in the store changes", () => {
    useAuthStore.setState({ role: "observer" });
    const { result, rerender } = renderHook(() =>
      useHasPermission("DeviceRemove"),
    );
    expect(result.current).toBe(false);

    useAuthStore.setState({ role: "administrator" });
    rerender();
    expect(result.current).toBe(true);
  });
});
