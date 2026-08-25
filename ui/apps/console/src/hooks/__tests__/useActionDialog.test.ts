import { describe, it, expect, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useActionDialog } from "../useActionDialog";

const entity = { uid: "uid-1", name: "my-device" };

describe("useActionDialog", () => {
  it("starts with no action", () => {
    const { result } = renderHook(() => useActionDialog());
    expect(result.current.action).toBeUndefined();
    expect(result.current.actionKey).toBe("closed");
  });

  it("requestAction sets the action", () => {
    const { result } = renderHook(() => useActionDialog());
    act(() => result.current.requestAction(entity, "accept"));
    expect(result.current.action).toEqual({ entity, operation: "accept" });
    expect(result.current.actionKey).toBe("accept/uid-1");
  });

  it("close clears the action", () => {
    const { result } = renderHook(() => useActionDialog());
    act(() => result.current.requestAction(entity, "remove"));
    act(() => result.current.close());
    expect(result.current.action).toBeUndefined();
    expect(result.current.actionKey).toBe("closed");
  });

  it("handleSuccess forwards to onSuccess", () => {
    const onSuccess = vi.fn();
    const { result } = renderHook(() => useActionDialog({ onSuccess }));
    act(() => result.current.handleSuccess("remove"));
    expect(onSuccess).toHaveBeenCalledWith("remove");
  });

  it("handleSuccess uses the latest onSuccess after rerender", () => {
    const first = vi.fn();
    const second = vi.fn();
    const { result, rerender } = renderHook(
      ({ onSuccess }) => useActionDialog({ onSuccess }),
      { initialProps: { onSuccess: first } },
    );
    rerender({ onSuccess: second });
    act(() => result.current.handleSuccess("accept"));
    expect(first).not.toHaveBeenCalled();
    expect(second).toHaveBeenCalledWith("accept");
  });

  it("requestAction and close are referentially stable", () => {
    const { result, rerender } = renderHook(() => useActionDialog());
    const { requestAction, close, handleSuccess } = result.current;
    act(() => result.current.requestAction(entity, "accept"));
    rerender();
    expect(result.current.requestAction).toBe(requestAction);
    expect(result.current.close).toBe(close);
    expect(result.current.handleSuccess).toBe(handleSuccess);
  });
});
