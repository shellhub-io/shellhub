import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { mockAccessPolicy, mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import type { AccessPolicy } from "@/client";
import AccessPolicies from "../index";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    listAccessPolicies: vi.fn(),
    deleteAccessPolicy: vi.fn(),
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
    listServiceAccounts: vi.fn(),
  }),
);

vi.mock("../AccessPolicyDrawer", () => ({
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="policy-drawer" /> : null,
}));

function renderList(policies: AccessPolicy[]) {
  sdk.listAccessPolicies.mockResolvedValue(mockSdkResponse(policies));
  return render(<AccessPolicies />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(mockNamespace()));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
  sdk.listServiceAccounts.mockResolvedValue(mockSdkResponse([]));
  sdk.deleteAccessPolicy.mockResolvedValue(mockSdkResponse(undefined));
});

describe("AccessPolicies", () => {
  it("puts the row actions in a menu without opening the policy behind it", async () => {
    const user = userEvent.setup();
    renderList([mockAccessPolicy({ id: "p1", name: "all" })]);

    await user.click(
      await screen.findByRole("button", { name: /actions for all/i }),
    );

    const menu = screen.getByRole("menu");
    expect(
      within(menu).getByRole("menuitem", { name: /edit/i }),
    ).toBeInTheDocument();
    expect(
      within(menu).getByRole("menuitem", { name: /delete/i }),
    ).toBeInTheDocument();

    expect(screen.queryByTestId("policy-drawer")).not.toBeInTheDocument();

    await user.keyboard("{Escape}");
    await user.click(screen.getByText("all"));
    expect(await screen.findByTestId("policy-drawer")).toBeInTheDocument();
  });
});
