import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { createTestWrapper, renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";

const mockInvalidate = vi.fn();

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getCustomer: vi.fn(),
    getSubscription: vi.fn(),
    createCustomer: vi.fn(),
    createSubscription: vi.fn(),
    attachPaymentMethod: vi.fn(),
    detachPaymentMethod: vi.fn(),
    setDefaultPaymentMethod: vi.fn(),
    createBillingPortalSession: vi.fn(),
  }),
);

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

async function importHooks() {
  return await import("../useBilling");
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useBilling mutations", () => {
  it("invalidates billing queries on customer creation", async () => {
    sdk.createCustomer.mockResolvedValue(mockSdkResponse(undefined));
    const { useCreateCustomer } = await importHooks();

    const { result } = renderHookWithClient(() => useCreateCustomer());

    await act(() => result.current.mutateAsync({}));

    await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    expect(sdk.createCustomer).toHaveBeenCalled();
  });

  it("invalidates billing queries on subscription creation", async () => {
    sdk.createSubscription.mockResolvedValue(mockSdkResponse(undefined));
    const { useCreateSubscription } = await importHooks();

    const { result } = renderHookWithClient(() => useCreateSubscription());

    await act(() => result.current.mutateAsync({}));

    await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
  });

  it("propagates 402 errors from subscription creation", async () => {
    const err = Object.assign(new Error("payment required"), {
      isAxiosError: true,
      response: { status: 402 },
    });
    sdk.createSubscription.mockRejectedValue(err);
    const { useCreateSubscription } = await importHooks();

    const { result } = renderHookWithClient(() => useCreateSubscription());

    await expect(result.current.mutateAsync({})).rejects.toBe(err);
    expect(mockInvalidate).not.toHaveBeenCalled();
  });

  it("attach/detach/default run through the SDK mutations", async () => {
    sdk.attachPaymentMethod.mockResolvedValue(mockSdkResponse(undefined));
    sdk.detachPaymentMethod.mockResolvedValue(mockSdkResponse(undefined));
    sdk.setDefaultPaymentMethod.mockResolvedValue(mockSdkResponse(undefined));
    const {
      useAttachPaymentMethod,
      useDetachPaymentMethod,
      useSetDefaultPaymentMethod,
    } = await importHooks();

    const wrapper = createTestWrapper();
    const attachHook = renderHook(() => useAttachPaymentMethod(), { wrapper });
    await act(() =>
      attachHook.result.current.mutateAsync({ body: { id: "pm_1" } }),
    );

    const detachHook = renderHook(() => useDetachPaymentMethod(), { wrapper });
    await act(() =>
      detachHook.result.current.mutateAsync({ body: { id: "pm_1" } }),
    );

    const defHook = renderHook(() => useSetDefaultPaymentMethod(), { wrapper });
    await act(() =>
      defHook.result.current.mutateAsync({ body: { id: "pm_1" } }),
    );

    expect(sdk.attachPaymentMethod).toHaveBeenCalled();
    expect(sdk.detachPaymentMethod).toHaveBeenCalled();
    expect(sdk.setDefaultPaymentMethod).toHaveBeenCalled();
  });
});

describe("useCreateSubscription (query key coverage)", () => {
  it("calls the mutation fn and then invalidates on success", async () => {
    sdk.createSubscription.mockResolvedValue(mockSdkResponse(undefined));
    const { useCreateSubscription } = await importHooks();

    const { result } = renderHookWithClient(() => useCreateSubscription());

    await act(() => result.current.mutateAsync({}));

    await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    expect(sdk.createSubscription).toHaveBeenCalled();
  });
});

describe("useCustomer", () => {
  it("does not call the SDK when enabled=false", async () => {
    const { useCustomer } = await importHooks();

    renderHookWithClient(() => useCustomer(false));

    expect(sdk.getCustomer).not.toHaveBeenCalled();
  });

  it("returns undefined customer when the query has no data", async () => {
    const { useCustomer } = await importHooks();

    const { result } = renderHookWithClient(() => useCustomer(false));

    expect(result.current.customer).toBeUndefined();
  });
});

describe("useSubscription", () => {
  it("does not call the SDK when enabled=false", async () => {
    const { useSubscription } = await importHooks();

    renderHookWithClient(() => useSubscription(false));

    expect(sdk.getSubscription).not.toHaveBeenCalled();
  });

  it("exposes a refetch function even when disabled", async () => {
    const { useSubscription } = await importHooks();

    const { result } = renderHookWithClient(() => useSubscription(false));

    expect(typeof result.current.refetch).toBe("function");
  });

  it("returns undefined subscription when query has no data", async () => {
    const { useSubscription } = await importHooks();

    const { result } = renderHookWithClient(() => useSubscription(false));

    expect(result.current.subscription).toBeUndefined();
  });
});

describe("useOpenBillingPortal", () => {
  it("opens the URL the billing portal route returns", async () => {
    const openSpy = vi.spyOn(window, "open").mockReturnValue(null);
    sdk.createBillingPortalSession.mockResolvedValue(
      mockSdkResponse({ url: "https://billing.stripe.com/session/abc" }),
    );
    const { useOpenBillingPortal } = await importHooks();

    const { result } = renderHookWithClient(() => useOpenBillingPortal());

    await act(() => result.current.mutateAsync());

    expect(sdk.createBillingPortalSession).toHaveBeenCalled();
    expect(openSpy).toHaveBeenCalledWith(
      "https://billing.stripe.com/session/abc",
      "_blank",
      "noopener,noreferrer",
    );
    openSpy.mockRestore();
  });

  it("throws when the response is missing a URL", async () => {
    sdk.createBillingPortalSession.mockResolvedValue(mockSdkResponse({}));
    const { useOpenBillingPortal } = await importHooks();

    const { result } = renderHookWithClient(() => useOpenBillingPortal());

    await expect(result.current.mutateAsync()).rejects.toThrow(/portal URL/i);
  });
});
