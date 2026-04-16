import { describe, it, expect, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useAuthStore } from "@/stores/authStore";
import { useHasPermission } from "../useHasPermission";

beforeEach(() => {
  useAuthStore.setState({ role: null });
});

describe("useHasPermission", () => {
  it("returns false when role is null", () => {
    useAuthStore.setState({ role: null });
    const { result } = renderHook(() => useHasPermission("device:connect"));
    expect(result.current).toBe(false);
  });

  it("returns true for observer on observer-level action", () => {
    useAuthStore.setState({ role: "observer" });
    const { result } = renderHook(() => useHasPermission("device:connect"));
    expect(result.current).toBe(true);
  });

  it("returns false for observer on operator-level action", () => {
    useAuthStore.setState({ role: "observer" });
    const { result } = renderHook(() => useHasPermission("device:accept"));
    expect(result.current).toBe(false);
  });

  it("returns true for administrator on administrator-level action", () => {
    useAuthStore.setState({ role: "administrator" });
    const { result } = renderHook(() => useHasPermission("publicKey:create"));
    expect(result.current).toBe(true);
  });

  it("returns false for administrator on owner-level action", () => {
    useAuthStore.setState({ role: "administrator" });
    const { result } = renderHook(() => useHasPermission("namespace:delete"));
    expect(result.current).toBe(false);
  });

  it("returns true for owner on all actions", () => {
    useAuthStore.setState({ role: "owner" });
    const { result: r1 } = renderHook(() => useHasPermission("namespace:delete"));
    const { result: r2 } = renderHook(() => useHasPermission("billing:subscribe"));
    expect(r1.current).toBe(true);
    expect(r2.current).toBe(true);
  });

  it("re-evaluates when role changes in the store", () => {
    useAuthStore.setState({ role: "observer" });
    const { result, rerender } = renderHook(() => useHasPermission("device:remove"));
    expect(result.current).toBe(false);

    useAuthStore.setState({ role: "administrator" });
    rerender();
    expect(result.current).toBe(true);
  });
});
