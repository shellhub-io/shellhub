import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { createTestWrapper, renderHookWithClient } from "@/tests/wrapper";

const mockCreateCustomer = vi.fn();
const mockCreateSubscription = vi.fn();
const mockAttach = vi.fn();
const mockDetach = vi.fn();
const mockSetDefault = vi.fn();
const mockGetCustomerFn = vi.fn();
const mockGetSubscriptionFn = vi.fn();
const mockInvalidate = vi.fn();
const mockCreateBillingPortalSession = vi.fn();

vi.mock("@/client/@tanstack/react-query.gen", () => ({
  getCustomerOptions: vi.fn(() => ({
    queryKey: [{ _id: "getCustomer" }],
    queryFn: mockGetCustomerFn,
  })),
  getSubscriptionOptions: vi.fn(() => ({
    queryKey: [{ _id: "getSubscription" }],
    queryFn: mockGetSubscriptionFn,
  })),
  createCustomerMutation: vi.fn(() => ({ mutationFn: mockCreateCustomer })),
  createSubscriptionMutation: vi.fn(() => ({
    mutationFn: mockCreateSubscription,
  })),
  attachPaymentMethodMutation: vi.fn(() => ({ mutationFn: mockAttach })),
  detachPaymentMethodMutation: vi.fn(() => ({ mutationFn: mockDetach })),
  setDefaultPaymentMethodMutation: vi.fn(() => ({
    mutationFn: mockSetDefault,
  })),
}));

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

vi.mock("@/client", () => ({
  createBillingPortalSession: (): unknown => mockCreateBillingPortalSession(),
}));

async function importHooks() {
  return await import("../useBilling");
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useBilling mutations", () => {
  it("invalidates billing queries on customer creation", async () => {
    mockCreateCustomer.mockResolvedValue(undefined);
    const { useCreateCustomer } = await importHooks();

    const { result } = renderHookWithClient(() => useCreateCustomer());

    await act(() => result.current.mutateAsync({}));

    await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    expect(mockCreateCustomer).toHaveBeenCalled();
  });

  it("invalidates billing queries on subscription creation", async () => {
    mockCreateSubscription.mockResolvedValue(undefined);
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
    mockCreateSubscription.mockRejectedValue(err);
    const { useCreateSubscription } = await importHooks();

    const { result } = renderHookWithClient(() => useCreateSubscription());

    await expect(result.current.mutateAsync({})).rejects.toBe(err);
    expect(mockInvalidate).not.toHaveBeenCalled();
  });

  it("attach/detach/default run through the SDK mutations", async () => {
    mockAttach.mockResolvedValue(undefined);
    mockDetach.mockResolvedValue(undefined);
    mockSetDefault.mockResolvedValue(undefined);
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

    expect(mockAttach).toHaveBeenCalled();
    expect(mockDetach).toHaveBeenCalled();
    expect(mockSetDefault).toHaveBeenCalled();
  });
});

describe("useCreateSubscription (query key coverage)", () => {
  it("calls the mutation fn and then invalidates on success", async () => {
    mockCreateSubscription.mockResolvedValue(undefined);
    const { useCreateSubscription } = await importHooks();

    const { result } = renderHookWithClient(() => useCreateSubscription());

    await act(() => result.current.mutateAsync({}));

    await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    expect(mockCreateSubscription).toHaveBeenCalled();
  });
});

describe("useCustomer", () => {
  it("does not call the queryFn when enabled=false", async () => {
    const { useCustomer } = await importHooks();

    renderHookWithClient(() => useCustomer(false));

    expect(mockGetCustomerFn).not.toHaveBeenCalled();
  });

  it("returns undefined customer when the query has no data", async () => {
    const { useCustomer } = await importHooks();

    const { result } = renderHookWithClient(() => useCustomer(false));

    expect(result.current.customer).toBeUndefined();
  });
});

describe("useSubscription", () => {
  it("does not call the queryFn when enabled=false", async () => {
    const { useSubscription } = await importHooks();

    renderHookWithClient(() => useSubscription(false));

    expect(mockGetSubscriptionFn).not.toHaveBeenCalled();
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
    mockCreateBillingPortalSession.mockResolvedValue({
      data: { url: "https://billing.stripe.com/session/abc" },
    });
    const { useOpenBillingPortal } = await importHooks();

    const { result } = renderHookWithClient(() => useOpenBillingPortal());

    await act(() => result.current.mutateAsync());

    expect(mockCreateBillingPortalSession).toHaveBeenCalled();
    expect(openSpy).toHaveBeenCalledWith(
      "https://billing.stripe.com/session/abc",
      "_blank",
      "noopener,noreferrer",
    );
    openSpy.mockRestore();
  });

  it("throws when the response is missing a URL", async () => {
    mockCreateBillingPortalSession.mockResolvedValue({ data: {} });
    const { useOpenBillingPortal } = await importHooks();

    const { result } = renderHookWithClient(() => useOpenBillingPortal());

    await expect(result.current.mutateAsync()).rejects.toThrow(/portal URL/i);
  });
});
