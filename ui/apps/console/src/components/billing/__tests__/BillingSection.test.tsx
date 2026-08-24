import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

const mockLocation = { hash: "", pathname: "/settings", search: "" };

vi.mock("react-router-dom", async () => {
  const actual =
    await vi.importActual<typeof import("react-router-dom")>(
      "react-router-dom",
    );
  return { ...actual, useLocation: () => mockLocation };
});

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getSubscription: vi.fn(),
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
    createSubscription: vi.fn(),
    createBillingPortalSession: vi.fn(),
  }),
);

vi.mock("@/api/errors", () => ({
  isSdkError: (err: unknown): boolean =>
    typeof err === "object" && err !== null && "status" in err,
}));

vi.mock("@/components/common/BaseDialog", async () => ({
  default: (await import("@/tests/mocks")).MockBaseDialog,
}));

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

function renderSection() {
  return render(
    <React.Suspense fallback={null}>
      <BillingSection sectionId="billing" />
    </React.Suspense>,
    { wrapper: createTestWrapper({ initialEntries: ["/"] }) },
  );
}

function billingRecord(active: boolean) {
  return {
    active,
    status: "active" as const,
    customer_id: "cus_123",
    subscription_id: "sub_123",
    current_period_end: 0,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  };
}

function setStatus(
  status: string,
  extra: { end_at?: number; invoices?: unknown[] } = {},
) {
  sdk.getNamespace.mockResolvedValue(
    mockSdkResponse(mockNamespace({ billing: billingRecord(true) })),
  );
  sdk.getSubscription.mockResolvedValue(mockSdkResponse({ status, ...extra }));
}

function setInactive() {
  sdk.getNamespace.mockResolvedValue(
    mockSdkResponse(mockNamespace({ billing: null })),
  );
  sdk.getSubscription.mockResolvedValue(mockSdkResponse(null));
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
  sdk.createSubscription.mockResolvedValue(mockSdkResponse(undefined));
  sdk.createBillingPortalSession.mockResolvedValue(
    mockSdkResponse({ url: "https://billing.stripe.com/session" }),
  );
  setInactive();
});

describe("BillingSection — Subscribe button visibility", () => {
  it("shows Subscribe button when status is 'inactive'", async () => {
    setInactive();
    renderSection();
    expect(
      await screen.findByRole("button", { name: /subscribe/i }),
    ).toBeInTheDocument();
  });

  it("does not show portal button when status is 'inactive'", async () => {
    setInactive();
    renderSection();
    await screen.findByRole("button", { name: /subscribe/i });
    expect(
      screen.queryByRole("button", { name: /open portal/i }),
    ).not.toBeInTheDocument();
  });

  it("shows Subscribe button when status is 'canceled'", async () => {
    setStatus("canceled");
    renderSection();
    expect(
      await screen.findByRole("button", { name: /subscribe/i }),
    ).toBeInTheDocument();
  });

  it("shows Subscribe button when status is 'incomplete_expired'", async () => {
    setStatus("incomplete_expired");
    renderSection();
    expect(
      await screen.findByRole("button", { name: /subscribe/i }),
    ).toBeInTheDocument();
  });

  it("shows portal button when status is 'canceled'", async () => {
    setStatus("canceled");
    renderSection();
    expect(
      await screen.findByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("shows portal button when status is 'incomplete_expired'", async () => {
    setStatus("incomplete_expired");
    renderSection();
    expect(
      await screen.findByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'active'", async () => {
    setStatus("active");
    renderSection();
    await screen.findByRole("button", { name: /open portal/i });
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("shows portal button when status is 'active'", async () => {
    setStatus("active");
    renderSection();
    expect(
      await screen.findByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'incomplete'", async () => {
    setStatus("incomplete");
    renderSection();
    await screen.findByRole("button", { name: /open portal/i });
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("shows portal button when status is 'incomplete'", async () => {
    setStatus("incomplete");
    renderSection();
    expect(
      await screen.findByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'unpaid'; shows portal instead", async () => {
    setStatus("unpaid");
    renderSection();
    await screen.findByRole("button", { name: /open portal/i });
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'paused'", async () => {
    setStatus("paused");
    renderSection();
    await screen.findByRole("button", { name: /open portal/i });
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("shows portal button when status is 'paused'", async () => {
    setStatus("paused");
    renderSection();
    expect(
      await screen.findByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });

  it("does not show Subscribe button when status is 'past_due'", async () => {
    setStatus("past_due");
    renderSection();
    await screen.findByRole("button", { name: /open portal/i });
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("shows portal button when status is 'past_due'", async () => {
    setStatus("past_due");
    renderSection();
    expect(
      await screen.findByRole("button", { name: /open portal/i }),
    ).toBeInTheDocument();
  });
});

describe("BillingSection — non-owner", () => {
  beforeEach(() => {
    seedAuthStore({ role: "administrator" });
  });

  it("shows the 'Owner-only' row", async () => {
    renderSection();
    expect(await screen.findByText("Owner-only")).toBeInTheDocument();
  });

  it("does not show the Subscribe button", async () => {
    renderSection();
    await screen.findByText("Owner-only");
    expect(
      screen.queryByRole("button", { name: /subscribe/i }),
    ).not.toBeInTheDocument();
  });

  it("does not show the portal button", async () => {
    setStatus("active");
    renderSection();
    await screen.findByText("Owner-only");
    expect(
      screen.queryByRole("button", { name: /open portal/i }),
    ).not.toBeInTheDocument();
  });

  it("does not show banners", async () => {
    setStatus("past_due");
    renderSection();
    await screen.findByText("Owner-only");
    expect(screen.queryByText(/payment overdue/i)).not.toBeInTheDocument();
  });
});

describe("BillingSection — banners", () => {
  it("shows 'Payment overdue' banner for 'past_due' status", async () => {
    setStatus("past_due");
    renderSection();
    expect(await screen.findByText("Payment overdue")).toBeInTheDocument();
  });

  it("shows 'Subscription incomplete' banner with billing-portal wording for 'incomplete'", async () => {
    setStatus("incomplete");
    renderSection();
    expect(
      await screen.findByText("Subscription incomplete"),
    ).toBeInTheDocument();
    expect(screen.getByText(/open the billing portal/i)).toBeInTheDocument();
  });

  it("shows 'Subscription expired' banner with 'Subscribe again' wording for 'incomplete_expired'", async () => {
    setStatus("incomplete_expired");
    renderSection();
    expect(await screen.findByText("Subscription expired")).toBeInTheDocument();
    expect(screen.getByText(/subscribe again/i)).toBeInTheDocument();
  });

  it("shows no banner for 'active' status", async () => {
    setStatus("active");
    renderSection();
    await screen.findByRole("button", { name: /open portal/i });
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });
});

describe("BillingSection — Subscribe button interaction", () => {
  it("clicking Subscribe opens BillingDialog", async () => {
    const user = userEvent.setup();
    setInactive();
    renderSection();
    await user.click(await screen.findByRole("button", { name: /subscribe/i }));
    expect(await screen.findByTestId("billing-letter")).toBeInTheDocument();
    expect(
      screen.getByRole("dialog", { name: /subscribe to shellhub cloud/i }),
    ).toBeInTheDocument();
  });

  it("closing BillingDialog via the X button hides it", async () => {
    const user = userEvent.setup();
    setInactive();
    renderSection();
    await user.click(await screen.findByRole("button", { name: /subscribe/i }));
    await screen.findByTestId("billing-letter");
    await user.click(screen.getByRole("button", { name: /close wizard/i }));
    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
    );
  });
});

describe("BillingSection — status badge", () => {
  it("shows 'Inactive' badge when there is no subscription", async () => {
    setInactive();
    renderSection();
    expect(await screen.findByText("Inactive")).toBeInTheDocument();
  });

  it("shows 'Active' badge when status is active", async () => {
    setStatus("active");
    renderSection();
    expect(await screen.findByText("Active")).toBeInTheDocument();
  });

  it("shows 'Past due' badge when status is past_due", async () => {
    setStatus("past_due");
    renderSection();
    expect(await screen.findByText("Past due")).toBeInTheDocument();
  });
});
