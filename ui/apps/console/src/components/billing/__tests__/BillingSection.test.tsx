import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
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

function setStatus(
  status: string,
  extra: { end_at?: number; invoices?: unknown[] } = {},
) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(
        mockNamespace({
          billing: {
            customer_id: "cus_123",
            subscription: {
              id: "sub_123",
              status: "active" as const,
              current_period_end: 0,
            },
            created_at: "2024-01-01T00:00:00Z",
            updated_at: "2024-01-01T00:00:00Z",
          },
        }),
      ),
    ),
    http.get("*/api/billing/subscription", () =>
      HttpResponse.json({ status, ...extra }),
    ),
  );
}

function setInactive() {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace({ billing: null })),
    ),
    http.get("*/api/billing/subscription", () => HttpResponse.json(null)),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  server.use(
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
    ),
    http.post(
      "*/api/billing/subscription",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.post("*/api/billing/portal", () =>
      HttpResponse.json({ url: "https://billing.stripe.com/session" }),
    ),
  );
  setInactive();
});

const SUBSCRIBE = /subscribe/i;
const PORTAL = /open portal/i;

describe("BillingSection — Subscribe / portal button visibility", () => {
  it.each([
    ["inactive", true, false],
    ["canceled", true, true],
    ["incomplete_expired", true, true],
    ["active", false, true],
    ["incomplete", false, true],
    ["unpaid", false, true],
    ["paused", false, true],
    ["past_due", false, true],
  ] as const)(
    "status %s offers subscribe=%s portal=%s",
    async (status, subscribe, portal) => {
      if (status === "inactive") setInactive();
      else setStatus(status);
      renderSection();

      await waitFor(() => {
        expect(screen.queryByRole("button", { name: SUBSCRIBE }) !== null).toBe(
          subscribe,
        );
        expect(screen.queryByRole("button", { name: PORTAL }) !== null).toBe(
          portal,
        );
      });
    },
  );
});

describe("BillingSection — non-owner", () => {
  beforeEach(() => {
    seedAuthStore({ role: "administrator" });
  });

  it("shows the 'Owner-only' row instead of the Subscribe button", async () => {
    setInactive();
    renderSection();
    await screen.findByText("Owner-only");
    expect(
      screen.queryByRole("button", { name: SUBSCRIBE }),
    ).not.toBeInTheDocument();
  });

  it("shows neither the portal button nor the banners", async () => {
    setStatus("past_due");
    renderSection();
    await screen.findByText("Owner-only");
    expect(
      screen.queryByRole("button", { name: PORTAL }),
    ).not.toBeInTheDocument();
    expect(screen.queryByText(/payment overdue/i)).not.toBeInTheDocument();
  });
});

describe("BillingSection — banners", () => {
  it.each([
    ["past_due", "Payment overdue", /open the billing portal/i],
    ["incomplete", "Subscription incomplete", /open the billing portal/i],
    ["incomplete_expired", "Subscription expired", /subscribe again/i],
  ] as const)(
    "status %s shows the '%s' banner",
    async (status, title, wording) => {
      setStatus(status);
      renderSection();
      expect(await screen.findByText(title)).toBeInTheDocument();
      expect(screen.getByText(wording)).toBeInTheDocument();
    },
  );

  it("shows no banner for 'active' status", async () => {
    setStatus("active");
    renderSection();
    await screen.findByRole("button", { name: PORTAL });
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });
});

describe("BillingSection — Subscribe button interaction", () => {
  it("clicking Subscribe opens BillingDialog", async () => {
    const user = userEvent.setup();
    setInactive();
    renderSection();
    await user.click(await screen.findByRole("button", { name: SUBSCRIBE }));
    expect(await screen.findByTestId("billing-letter")).toBeInTheDocument();
    expect(
      screen.getByRole("dialog", { name: /subscribe to shellhub cloud/i }),
    ).toBeInTheDocument();
  });
});

describe("BillingSection — status badge", () => {
  it.each([
    ["inactive", "Inactive"],
    ["active", "Active"],
    ["past_due", "Past due"],
  ] as const)("status %s shows the '%s' badge", async (status, badge) => {
    if (status === "inactive") setInactive();
    else setStatus(status);
    renderSection();
    expect(await screen.findByText(badge)).toBeInTheDocument();
  });
});
