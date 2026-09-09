import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import AdminFirewallRules from "../index";
import { createTestWrapper } from "@/tests/wrapper";
import { mockFirewallRule } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

function setRules(
  rules: ReturnType<typeof mockFirewallRule>[],
  total?: number,
) {
  server.use(
    http.get("*/admin/api/firewall/rules", () =>
      jsonWithTotal(rules, total ?? rules.length),
    ),
  );
}

function searchbox() {
  return screen.getByRole("searchbox", {
    name: "Search firewall rules by action, priority, IP, or username",
  });
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminFirewallRules />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("AdminFirewallRules", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ isAdmin: true });
    setRules([]);
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading firewall rules..." text', () => {
      server.use(
        http.get("*/admin/api/firewall/rules", () => new Promise(() => {})),
      );
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading firewall rules...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No firewall rules found" when the list is empty', async () => {
      renderPage();
      expect(
        await screen.findByText("No firewall rules found"),
      ).toBeInTheDocument();
    });
  });

  describe("rule rows", () => {
    it("renders a row for each returned rule", async () => {
      setRules(
        [
          mockFirewallRule({ id: "r1", priority: 1 }),
          mockFirewallRule({ id: "r2", priority: 2 }),
        ],
        2,
      );
      renderPage();
      await waitFor(() => expect(screen.getAllByText("Allow").length).toBe(2));
      expect(screen.getAllByText("1")[0]).toBeInTheDocument();
      expect(screen.getAllByText("2")[0]).toBeInTheDocument();
    });

    it("renders the rule's own action, IP, username and state", async () => {
      setRules([
        mockFirewallRule({
          action: "allow",
          source_ip: "192.168.1.0/24",
          username: "alice",
          active: true,
        }),
      ]);
      renderPage();
      expect(await screen.findByText("Allow")).toBeInTheDocument();
      expect(screen.getByText("192.168.1.0/24")).toBeInTheDocument();
      expect(screen.getByText("alice")).toBeInTheDocument();
      expect(screen.getByText("Active")).toBeInTheDocument();
    });

    it("renders wildcard source_ip and username as 'Any IP' and 'All users'", async () => {
      setRules([
        mockFirewallRule({
          action: "deny",
          source_ip: ".*",
          username: ".*",
          active: false,
        }),
      ]);
      renderPage();
      expect(await screen.findByText("Deny")).toBeInTheDocument();
      expect(screen.getByText("Any IP")).toBeInTheDocument();
      expect(screen.getByText("All users")).toBeInTheDocument();
      expect(screen.getByText("Inactive")).toBeInTheDocument();
    });

    it("navigates to the detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      setRules([mockFirewallRule({ id: "rule-abc", priority: 99 })]);
      renderPage();

      await user.click(await screen.findByText("99"));
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/firewall-rules/rule-abc",
      );
    });

    it("renders the tenant_id as a namespace link", async () => {
      setRules([mockFirewallRule({ tenant_id: "tenant-xyz" })]);
      renderPage();
      const link = await screen.findByRole("link", { name: "tenant-xyz" });
      expect(link).toHaveAttribute("href", "/admin/namespaces/tenant-xyz");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the SDK returns an error", async () => {
      server.use(
        http.get("*/admin/api/firewall/rules", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderPage();
      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText("Something went wrong on our side. Try again."),
      ).toBeInTheDocument();
    });
  });

  describe("client-side search", () => {
    const allowRule = mockFirewallRule({
      id: "r1",
      action: "allow",
      priority: 5,
      source_ip: "172.16.0.1",
      username: ".*",
    });
    const denyRule = mockFirewallRule({
      id: "r2",
      action: "deny",
      priority: 777,
      source_ip: ".*",
      username: "zara",
    });

    beforeEach(() => {
      setRules([allowRule, denyRule]);
    });

    it.each([
      ["deny", "Deny", "Allow"],
      ["172.16.0.1", "172.16.0.1", "zara"],
      ["zara", "zara", "172.16.0.1"],
      ["777", "777", "Allow"],
    ])("searching '%s' keeps '%s' and drops '%s'", async (term, kept, gone) => {
      const user = userEvent.setup();
      renderPage();

      await screen.findByText(kept);

      await user.type(searchbox(), term);

      await waitFor(() =>
        expect(screen.queryByText(gone)).not.toBeInTheDocument(),
      );
      expect(screen.getByText(kept)).toBeInTheDocument();
    });

    it('shows "No rules matching" message when search has no results', async () => {
      const user = userEvent.setup();
      renderPage();

      await screen.findByText("Allow");

      await user.type(searchbox(), "zzz-no-match");

      await screen.findByText(/No rules matching/);
    });
  });
});
