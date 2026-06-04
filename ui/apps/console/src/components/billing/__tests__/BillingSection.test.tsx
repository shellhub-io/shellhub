import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const mockLocation = { hash: "", pathname: "/settings", search: "" };

vi.mock("react-router-dom", async () => {
  const actual =
    await vi.importActual<typeof import("react-router-dom")>(
      "react-router-dom",
    );
  return { ...actual, useLocation: () => mockLocation };
});

const mockOpenPortalMutate = vi.fn();
const mockSubscriptionData = {
  value: null as {
    status: string;
    end_at?: number;
    invoices?: unknown[];
  } | null,
};
const mockOpenPortalIsPending = { value: false };
const mockOpenPortalIsError = { value: false };

// Also includes useCreateSubscription because the real BillingDialog is
// lazy-loaded inside BillingSection — vi.mock() on the BillingDialog module
// does not intercept React.lazy dynamic imports, so the real component
// renders and needs all its hook dependencies satisfied.
vi.mock("@/hooks/useBilling", () => ({
  useSubscription: () => ({
    subscription: mockSubscriptionData.value,
    isLoading: false,
    refetch: vi.fn().mockResolvedValue({ data: { status: "active" } }),
  }),
  useOpenBillingPortal: () => ({
    mutate: (...args: unknown[]) => {
      mockOpenPortalMutate(...args);
    },
    isPending: mockOpenPortalIsPending.value,
    isError: mockOpenPortalIsError.value,
  }),
  useCreateSubscription: () => ({
    mutateAsync: vi.fn().mockResolvedValue(undefined),
    isPending: false,
  }),
}));

const mockNamespaceData = {
  value: null as { billing?: Record<string, unknown> } | null,
};

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: () => ({
    namespace: mockNamespaceData.value,
    isLoading: false,
  }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: () => ({ tenant: "tenant-id-123" }),
}));

const mockCanSubscribe = { value: true };

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => mockCanSubscribe.value,
}));

const mockInvalidate = vi.fn();

vi.mock("@/hooks/useInvalidateQueries", () => ({
  useInvalidateByIds: () => mockInvalidate,
}));

// Deep deps of the real BillingDialog (which gets rendered via React.lazy):
vi.mock("@/hooks/useFocusTrap", () => ({ useFocusTrap: vi.fn() }));

vi.mock("@/api/errors", () => ({
  isSdkError: (err: unknown): boolean =>
    typeof err === "object" && err !== null && "status" in err,
}));

// BaseDialog uses native <dialog> + focus trap — replace with a plain wrapper
vi.mock("@/components/common/BaseDialog", () => ({
  default: ({
    open,
    children,
    "aria-label": ariaLabel,
  }: {
    open: boolean;
    onClose: () => void;
    canClose?: () => boolean;
    children: React.ReactNode;
    "aria-label"?: string;
  }) => {
    if (!open) return null;
    return React.createElement(
      "div",
      { role: "dialog", "aria-label": ariaLabel },
      children,
    );
  },
}));

// Stub the wizard steps — we only care that the dialog opens/closes here.
vi.mock("@/components/billing/BillingLetter", () => ({
  default: () =>
    React.createElement("div", { "data-testid": "billing-letter" }),
}));
vi.mock("@/components/billing/BillingPayment", () => ({ default: () => null }));
vi.mock("@/components/billing/BillingCheckout", () => ({
  default: () => null,
}));
vi.mock("@/components/billing/BillingSuccessful", () => ({
  default: () => null,
}));

import BillingSection from "../BillingSection";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </MemoryRouter>
  );
}

function renderSection() {
  return render(
    <React.Suspense fallback={null}>
      <BillingSection sectionId="billing" />
    </React.Suspense>,
    { wrapper: createWrapper() },
  );
}

function setStatus(
  status: string,
  extra: { end_at?: number; invoices?: unknown[] } = {},
) {
  mockSubscriptionData.value = { status, ...extra };
  mockNamespaceData.value = { billing: { customer_id: "cus_123" } };
}

function setInactive() {
  mockSubscriptionData.value = null;
  mockNamespaceData.value = { billing: {} };
}

beforeEach(() => {
  vi.clearAllMocks();
  mockCanSubscribe.value = true;
  mockOpenPortalIsPending.value = false;
  mockOpenPortalIsError.value = false;
  setInactive();
});

afterEach(cleanup);

describe("BillingSection — Subscribe button visibility", () => {
  it("shows Subscribe button when status is 'inactive'", () => {
    setInactive();
    renderSection();
    expect(
      screen.getByRole("button", { name: /subscribe/i }),
    ).toBeInTheDocument();
  });

  it("does not show portal button when status is 'inactive'", () => {
    setInactive();
    renderSection();
    expect(
      screen.queryByRole("button", { name: /open portal/i }),
    ).not.toBeInTheDocument();
  });

  it("shows Subscribe button when status is 'canceled'", () => {
    setStatus("canceled");
    renderSection();
    expect(
      screen.getByRole("button", { name: /subscribe/i }),
    ).toBeInTheDocument();
  });

  it("shows Subscribe button when status is 'incomplete_expired'", () => {
    setStatus("incomplete_expired");
    renderSection();
    expect(
      screen.getByRole("button", { name: /subscribe/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'active'", () => {
    setStatus("active");
    renderSection();
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("shows portal button when status is 'active'", () => {
    setStatus("active");
    renderSection();
    expect(
      screen.getByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'incomplete'", () => {
    setStatus("incomplete");
    renderSection();
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("shows portal button when status is 'incomplete'", () => {
    setStatus("incomplete");
    renderSection();
    expect(
      screen.getByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'unpaid'; shows portal instead", () => {
    setStatus("unpaid");
    renderSection();
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'paused'", () => {
    setStatus("paused");
    renderSection();
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("shows portal button when status is 'paused'", () => {
    setStatus("paused");
    renderSection();
    expect(
      screen.getByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'past_due'", () => {
    setStatus("past_due");
    renderSection();
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("shows portal button when status is 'past_due'", () => {
    setStatus("past_due");
    renderSection();
    expect(
      screen.getByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });
});

describe("BillingSection — non-owner", () => {
  beforeEach(() => {
    mockCanSubscribe.value = false;
  });

  it("shows the 'Owner-only' row", () => {
    renderSection();
    expect(screen.getByText("Owner-only")).toBeInTheDocument();
  });

  it("does not show the Subscribe button", () => {
    renderSection();
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("does not show the portal button", () => {
    setStatus("active");
    renderSection();
    expect(
      screen.queryByRole("button", { name: /open portal/i }),
    ).not.toBeInTheDocument();
  });

  it("does not show banners", () => {
    setStatus("past_due");
    renderSection();
    expect(screen.queryByText(/payment overdue/i)).not.toBeInTheDocument();
  });
});

describe("BillingSection — banners", () => {
  it("shows 'Payment overdue' banner for 'past_due' status", () => {
    setStatus("past_due");
    renderSection();
    expect(screen.getByText("Payment overdue")).toBeInTheDocument();
  });

  it("shows 'Subscription incomplete' banner with billing-portal wording for 'incomplete'", () => {
    setStatus("incomplete");
    renderSection();
    expect(screen.getByText("Subscription incomplete")).toBeInTheDocument();
    expect(screen.getByText(/open the billing portal/i)).toBeInTheDocument();
  });

  it("shows 'Subscription expired' banner with 'Subscribe again' wording for 'incomplete_expired'", () => {
    setStatus("incomplete_expired");
    renderSection();
    expect(screen.getByText("Subscription expired")).toBeInTheDocument();
    expect(screen.getByText(/subscribe again/i)).toBeInTheDocument();
  });

  it("shows no banner for 'active' status", () => {
    setStatus("active");
    renderSection();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });
});

describe("BillingSection — Subscribe button interaction", () => {
  it("clicking Subscribe opens BillingDialog", async () => {
    const user = userEvent.setup();
    setInactive();
    renderSection();
    await user.click(screen.getByRole("button", { name: /subscribe/i }));
    // Real BillingDialog is rendered via React.lazy; on step 1 it shows BillingLetter.
    expect(await screen.findByTestId("billing-letter")).toBeInTheDocument();
    expect(
      screen.getByRole("dialog", { name: /subscribe to shellhub cloud/i }),
    ).toBeInTheDocument();
  });

  it("closing BillingDialog via the X button hides it", async () => {
    const user = userEvent.setup();
    setInactive();
    renderSection();
    await user.click(screen.getByRole("button", { name: /subscribe/i }));
    await screen.findByTestId("billing-letter");
    await user.click(screen.getByRole("button", { name: /close wizard/i }));
    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
    );
  });
});

describe("BillingSection — status badge", () => {
  it("shows 'Inactive' badge when there is no subscription", () => {
    setInactive();
    renderSection();
    expect(screen.getByText("Inactive")).toBeInTheDocument();
  });

  it("shows 'Active' badge when status is active", () => {
    setStatus("active");
    renderSection();
    expect(screen.getByText("Active")).toBeInTheDocument();
  });

  it("shows 'Past due' badge when status is past_due", () => {
    setStatus("past_due");
    renderSection();
    expect(screen.getByText("Past due")).toBeInTheDocument();
  });
});
