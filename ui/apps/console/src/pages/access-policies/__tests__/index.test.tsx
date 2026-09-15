import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockAccessPolicy, mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import type { AccessPolicy } from "@/client/model";
import AccessPolicies from "../index";

vi.mock("../AccessPolicyDrawer", () => ({
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="policy-drawer" /> : null,
}));

function renderList(policies: AccessPolicy[]) {
  server.use(
    http.get("*/api/access-policies", () => HttpResponse.json(policies)),
  );
  return render(<AccessPolicies />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
    ),
    http.get("*/api/service-accounts", () => HttpResponse.json([])),
    http.delete(
      "*/api/access-policies/:id",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
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
