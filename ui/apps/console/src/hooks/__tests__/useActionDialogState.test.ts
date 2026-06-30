import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import {
  useActionDialogState,
  type UseActionDialogStateOptions,
} from "../useActionDialogState";
import { useDeviceActions } from "../useDeviceActions";
import { useContainerActions } from "../useContainerActions";
import { defaultConfig, type ClientConfig } from "@/env";

// Only the wrapper hooks read getConfig().cloud; the core hook never touches env.
const mockGetConfig = vi.fn<() => ClientConfig>();
vi.mock("@/env", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/env")>()),
  getConfig: () => mockGetConfig(),
}));

const entity = { uid: "device-123", name: "Device 123" };

const setup = (options: Partial<UseActionDialogStateOptions> = {}) =>
  renderHook(() => useActionDialogState({ enableBillingWarning: false, ...options }));

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, cloud: false });
});

describe("useActionDialogState", () => {
  it("requestAction opens the dialog for an entity/action pair", () => {
    const { result } = setup();

    act(() => result.current.requestAction(entity, "remove"));

    expect(result.current.operation).toEqual({ entity, action: "remove" });
  });

  it("close clears the operation", () => {
    const { result } = setup();

    act(() => result.current.requestAction(entity, "remove"));
    act(() => result.current.close());

    expect(result.current.operation).toBeUndefined();
  });

  describe("billing warning", () => {
    it("exposes onBillingWarning and billingEnabled only when enabled", () => {
      const off = setup({ enableBillingWarning: false }).result.current;
      expect(off.onBillingWarning).toBeUndefined();
      expect(off.billingEnabled).toBe(false);

      const on = setup({ enableBillingWarning: true }).result.current;
      expect(on.onBillingWarning).toBeTypeOf("function");
      expect(on.billingEnabled).toBe(true);
    });

    it("opening the warning dismisses the action dialog", () => {
      const { result } = setup({ enableBillingWarning: true });

      act(() => result.current.requestAction(entity, "accept"));
      act(() => result.current.onBillingWarning!());

      expect(result.current.billingWarningOpen).toBe(true);
      expect(result.current.operation).toBeUndefined();
    });

    it("closeBillingWarning closes the warning", () => {
      const { result } = setup({ enableBillingWarning: true });

      act(() => result.current.onBillingWarning!());
      act(() => result.current.closeBillingWarning());

      expect(result.current.billingWarningOpen).toBe(false);
    });
  });

  describe("runSuccess", () => {
    it("invokes onSuccess with the passed action and leaves the dialog open for the caller to close", () => {
      const onSuccess = vi.fn();
      const { result } = setup({ onSuccess });

      act(() => result.current.requestAction(entity, "remove"));
      act(() => result.current.runSuccess("remove"));

      expect(onSuccess).toHaveBeenCalledOnce();
      expect(onSuccess).toHaveBeenCalledWith("remove");
      expect(result.current.operation).toEqual({ entity, action: "remove" });
    });

    it("fires onSuccess with the captured action even after close() (cancel-race)", () => {
      const onSuccess = vi.fn();
      const { result } = setup({ onSuccess });

      act(() => result.current.requestAction(entity, "remove"));
      act(() => result.current.close());
      act(() => result.current.runSuccess("remove"));

      expect(onSuccess).toHaveBeenCalledOnce();
      expect(onSuccess).toHaveBeenCalledWith("remove");
    });

    it("reports the captured action, not whatever operation is currently live", () => {
      const onSuccess = vi.fn();
      const { result } = setup({ onSuccess });

      act(() => result.current.requestAction(entity, "accept"));
      act(() => result.current.requestAction(entity, "remove"));
      act(() => result.current.runSuccess("accept"));

      expect(onSuccess).toHaveBeenCalledWith("accept");
    });

    it("calls the latest onSuccess after the callback prop changes", () => {
      const first = vi.fn();
      const second = vi.fn();
      const { result, rerender } = renderHook(
        ({ onSuccess }: { onSuccess: UseActionDialogStateOptions["onSuccess"] }) =>
          useActionDialogState({ enableBillingWarning: false, onSuccess }),
        { initialProps: { onSuccess: first } },
      );

      act(() => result.current.requestAction(entity, "remove"));
      rerender({ onSuccess: second });
      act(() => result.current.runSuccess("remove"));

      expect(first).not.toHaveBeenCalled();
      expect(second).toHaveBeenCalledWith("remove");
    });
  });

  it("returns referentially stable callbacks across re-renders and state changes", () => {
    const { result, rerender } = setup({ onSuccess: vi.fn() });
    const callbacks = ["requestAction", "close", "closeBillingWarning", "runSuccess"] as const;
    const before = callbacks.map((name) => result.current[name]);

    act(() => result.current.requestAction(entity, "remove")); // operation changes
    rerender();

    callbacks.forEach((name, i) => expect(result.current[name]).toBe(before[i]));
  });
});

// Both wrappers are thin adapters over the core hook that only differ in their
// domain label, so they share one parametrised suite.
describe.each([
  { name: "useDeviceActions", useActions: useDeviceActions, target: { uid: "d-1", name: "my-device" } },
  { name: "useContainerActions", useActions: useContainerActions, target: { uid: "c-1", name: "my-container" } },
])("$name", ({ useActions, target }) => {
  it("derives onBillingWarning from getConfig().cloud", () => {
    mockGetConfig.mockReturnValue({ ...defaultConfig, cloud: true });
    expect(renderHook(() => useActions()).result.current.onBillingWarning).toBeTypeOf("function");

    mockGetConfig.mockReturnValue({ ...defaultConfig, cloud: false });
    expect(renderHook(() => useActions()).result.current.onBillingWarning).toBeUndefined();
  });

  it("stores a requested action as an operation", () => {
    const { result } = renderHook(() => useActions());

    act(() => result.current.requestAction(target, "accept"));

    expect(result.current.operation).toEqual({ entity: target, action: "accept" });
  });

  it("forwards onSuccess, even after close() (cancel-race)", () => {
    const onSuccess = vi.fn();
    const { result } = renderHook(() => useActions({ onSuccess }));

    act(() => result.current.requestAction(target, "remove"));
    act(() => result.current.close());
    act(() => result.current.runSuccess("remove"));

    expect(onSuccess).toHaveBeenCalledOnce();
    expect(onSuccess).toHaveBeenCalledWith("remove");
  });
});
