import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { mockNamespace } from "@/tests/factories";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import Team from "..";

beforeEach(() => {
  seedAuthStore({ role: "owner" });
  server.use(
    http.get("*/api/namespaces/api-key", () => jsonWithTotal([])),
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace({ members: [] })),
    ),
    http.get("*/api/namespaces/:tenant/members", () => jsonWithTotal([])),
    http.get("*/api/namespaces/:tenant/invitations", () => jsonWithTotal([])),
    http.get("*/api/access-policies", () => jsonWithTotal([])),
  );
});

function renderTeam() {
  render(<Team />, { wrapper: createTestWrapper({ initialEntries: ["/team"] }) });
}

describe("Team", () => {
  it("opens on the Members tab, its panel holding the members list", async () => {
    renderTeam();

    expect(screen.getByRole("tab", { name: "Members" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByRole("tabpanel", { name: "Members" })).toContainElement(
      await screen.findByText("No members yet"),
    );
  });

  it("swaps the panel to API keys when that tab is selected", async () => {
    const user = userEvent.setup();
    renderTeam();

    await user.click(screen.getByRole("tab", { name: "API Keys" }));

    const panel = screen.getByRole("tabpanel", { name: "API Keys" });
    expect(panel).toContainElement(await screen.findByText("No API keys yet"));
    expect(screen.getByRole("tab", { name: "API Keys" })).toHaveAttribute(
      "aria-controls",
      panel.id,
    );
  });

  it("keeps a labelled Members panel without mounting the members list when there is no tenant", () => {
    seedAuthStore({ role: "owner", tenant: null });
    renderTeam();

    const members = screen.getByRole("tab", { name: "Members" });
    expect(screen.getByRole("tabpanel", { name: "Members" }).id).toBe(
      members.getAttribute("aria-controls"),
    );
    expect(
      screen.queryByRole("button", { name: "Add Member" }),
    ).not.toBeInTheDocument();
  });
});
