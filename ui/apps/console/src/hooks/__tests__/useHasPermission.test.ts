import { describe, it, expect, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useAuthStore } from "@/stores/authStore";
import { useHasPermission } from "../useHasPermission";

beforeEach(() => {
  useAuthStore.setState({ role: null });
});

describe("useHasPermission", () => {
  it("re-evaluates when role changes in the store", () => {
    useAuthStore.setState({ role: "observer" });
    const { result, rerender } = renderHook(() =>
      useHasPermission("device:remove"),
    );
    expect(result.current).toBe(false);

    useAuthStore.setState({ role: "administrator" });
    rerender();
    expect(result.current).toBe(true);
  });
});
