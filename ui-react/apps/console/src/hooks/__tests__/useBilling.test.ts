import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act, cleanup } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

afterEach(() => {
  cleanup();
});

const mockCreateCustomer = vi.fn();
const mockCreateSubscription = vi.fn();
const mockAttach = vi.fn();
const mockDetach = vi.fn();
const mockSetDefault = vi.fn();
const mockGetCustomerFn = vi.fn();
const mockGetSubscriptionFn = vi.fn();
const mockInvalidate = vi.fn();
const mockAxiosPost = vi.fn();

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

vi.mock("@/api/client", () => ({
  default: {
    post: (...args: unknown[]) => {
      mockAxiosPost(...args);
      return mockAxiosPost.mock.results[mockAxiosPost.mock.results.length - 1]
        .value as unknown;
    },
  },
}));

async function importHooks() {
  return await import("../useBilling");
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
      queries: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useBilling mutations", () => {
  it("invalidates billing queries on customer creation", async () => {
    mockCreateCustomer.mockResolvedValue(undefined);
    const { useCreateCustomer } = await importHooks();

    const { result } = renderHook(() => useCreateCustomer(), {
      wrapper: createWrapper(),
    });

    await act(() => result.current.mutateAsync({}));

    await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    expect(mockCreateCustomer).toHaveBeenCalled();
  });

  it("invalidates billing queries on subscription creation", async () => {
    mockCreateSubscription.mockResolvedValue(undefined);
    const { useCreateSubscription } = await importHooks();

    const { result } = renderHook(() => useCreateSubscription(), {
      wrapper: createWrapper(),
    });

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

    const { result } = renderHook(() => useCreateSubscription(), {
      wrapper: createWrapper(),
    });

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

    const wrapper = createWrapper();
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

    const { result } = renderHook(() => useCreateSubscription(), {
      wrapper: createWrapper(),
    });

    await act(() => result.current.mutateAsync({}));

    await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    expect(mockCreateSubscription).toHaveBeenCalled();
  });
});

describe("useCustomer", () => {
  it("does not call the queryFn when enabled=false", async () => {
    const { useCustomer } = await importHooks();

    renderHook(() => useCustomer(false), { wrapper: createWrapper() });

    expect(mockGetCustomerFn).not.toHaveBeenCalled();
  });

  it("returns null customer when the query has no data", async () => {
    const { useCustomer } = await importHooks();

    const { result } = renderHook(() => useCustomer(false), {
      wrapper: createWrapper(),
    });

    expect(result.current.customer).toBeNull();
  });
});

describe("useSubscription", () => {
  it("does not call the queryFn when enabled=false", async () => {
    const { useSubscription } = await importHooks();

    renderHook(() => useSubscription(false), { wrapper: createWrapper() });

    expect(mockGetSubscriptionFn).not.toHaveBeenCalled();
  });

  it("exposes a refetch function even when disabled", async () => {
    const { useSubscription } = await importHooks();

    const { result } = renderHook(() => useSubscription(false), {
      wrapper: createWrapper(),
    });

    expect(typeof result.current.refetch).toBe("function");
  });

  it("returns null subscription when query has no data", async () => {
    const { useSubscription } = await importHooks();

    const { result } = renderHook(() => useSubscription(false), {
      wrapper: createWrapper(),
    });

    expect(result.current.subscription).toBeNull();
  });
});

describe("useOpenBillingPortal", () => {
  it("POSTs to /api/billing/portal and opens the returned URL", async () => {
    const openSpy = vi.spyOn(window, "open").mockReturnValue(null);
    mockAxiosPost.mockResolvedValue({
      data: { url: "https://billing.stripe.com/session/abc" },
    });
    const { useOpenBillingPortal } = await importHooks();

    const { result } = renderHook(() => useOpenBillingPortal(), {
      wrapper: createWrapper(),
    });

    await act(() => result.current.mutateAsync());

    expect(mockAxiosPost).toHaveBeenCalledWith("/api/billing/portal");
    expect(openSpy).toHaveBeenCalledWith(
      "https://billing.stripe.com/session/abc",
      "_blank",
      "noopener,noreferrer",
    );
    openSpy.mockRestore();
  });

  it("throws when the response is missing a URL", async () => {
    mockAxiosPost.mockResolvedValue({ data: {} });
    const { useOpenBillingPortal } = await importHooks();

    const { result } = renderHook(() => useOpenBillingPortal(), {
      wrapper: createWrapper(),
    });

    await expect(result.current.mutateAsync()).rejects.toThrow(/portal URL/i);
  });
});
