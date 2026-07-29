import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import AccessPolicies from "../index";
import type { AccessPolicy } from "@/client";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockPolicies =
  vi.fn<() => { policies: AccessPolicy[]; isLoading: boolean }>();

vi.mock("@/hooks/useAccessPolicies", () => ({
  useAccessPolicies: () => mockPolicies(),
}));

vi.mock("@/hooks/useAccessPolicyMutations", () => ({
  useDeleteAccessPolicy: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: () => ({ namespace: { name: "dev" } }),
}));

vi.mock("@/hooks/useServiceAccounts", () => ({
  useServiceAccounts: () => ({ serviceAccounts: [] }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: Object.assign(
    (selector?: (s: unknown) => unknown) => {
      const state = { tenant: "tenant1", userId: "user1" };
      return selector ? selector(state) : state;
    },
    { getState: () => ({}) },
  ),
}));

vi.mock("../AccessPolicyDrawer", () => ({
  // Standing in for the drawer, so opening it is observable without rendering it.
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="policy-drawer" /> : null,
}));

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function policy(overrides: Partial<AccessPolicy> = {}): AccessPolicy {
  return {
    id: "p1",
    name: "all",
    action: "allow",
    subject: { type: "all", value: "" },
    filter: { tags: [] },
    logins: ["*"],
    created_at: "2026-01-01T00:00:00.000Z",
    ...overrides,
  } as AccessPolicy;
}

function renderList(policies: AccessPolicy[]) {
  mockPolicies.mockReturnValue({ policies, isLoading: false });

  return render(
    <MemoryRouter>
      <AccessPolicies />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("AccessPolicies", () => {
  // One render for the whole story: rendering this page pulls in the real
  // floating-ui dropdown, and repeating it makes the parallel suite miss
  // timings elsewhere. The three facts are sequential anyway.
  it("puts the row actions in a menu without opening the policy behind it", async () => {
    const user = userEvent.setup();
    renderList([policy()]);

    await user.click(screen.getByRole("button", { name: /actions for all/i }));

    const menu = screen.getByRole("menu");
    expect(
      within(menu).getByRole("menuitem", { name: /edit/i }),
    ).toBeInTheDocument();
    expect(
      within(menu).getByRole("menuitem", { name: /delete/i }),
    ).toBeInTheDocument();

    // The row itself opens the policy, so the trigger has to swallow its own
    // click or the menu would open on top of the drawer it just opened.
    expect(screen.queryByTestId("policy-drawer")).not.toBeInTheDocument();

    await user.keyboard("{Escape}");
    await user.click(screen.getByText("all"));
    expect(await screen.findByTestId("policy-drawer")).toBeInTheDocument();
  });
});
