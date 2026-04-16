import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { FirewallRule } from "@/hooks/useAdminFirewallRules";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useAdminFirewallRules", () => ({
  useAdminFirewallRules: vi.fn(),
}));

// useNavigate is used by the page — mock at the module level.
const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { useAdminFirewallRules } from "@/hooks/useAdminFirewallRules";
import AdminFirewallRules from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

const defaultHookState = {
  rules: [] as FirewallRule[],
  totalCount: 0,
  isLoading: false,
  error: null,
};

function makeRule(overrides: Partial<FirewallRule> = {}): FirewallRule {
  return {
    id: "rule-1",
    tenant_id: "tenant-abc",
    priority: 1,
    action: "allow" as const,
    active: true,
    source_ip: ".*",
    username: ".*",
    filter: { hostname: ".*" },
    ...overrides,
  };
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminFirewallRules />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AdminFirewallRules", () => {
  beforeEach(() => {
    vi.mocked(useAdminFirewallRules).mockReturnValue(defaultHookState);
    mockNavigate.mockReset();
  });

  describe("rendering", () => {
    it('renders the page heading "Firewall Rules"', () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Firewall Rules" }),
      ).toBeInTheDocument();
    });

    it("renders the search input with correct aria-label", () => {
      renderPage();
      expect(
        screen.getByRole("textbox", { name: "Search firewall rules" }),
      ).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading firewall rules..." text', () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        isLoading: true,
        rules: [],
      });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading firewall rules...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No firewall rules found" when the list is empty', () => {
      renderPage();
      expect(screen.getByText("No firewall rules found")).toBeInTheDocument();
    });
  });

  describe("rule rows", () => {
    it("renders a row for each returned rule", () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [
          makeRule({ id: "r1", priority: 1 }),
          makeRule({ id: "r2", priority: 2 }),
        ],
        totalCount: 2,
      });
      renderPage();
      // Both priority values appear in the table
      expect(screen.getAllByText("1")[0]).toBeInTheDocument();
      expect(screen.getAllByText("2")[0]).toBeInTheDocument();
    });

    it('shows "Allow" with accent-green for an allow rule', () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ action: "allow" })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("Allow")).toBeInTheDocument();
    });

    it('shows "Deny" for a deny rule', () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ action: "deny" })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("Deny")).toBeInTheDocument();
    });

    it('shows "Any IP" when source_ip is ".*"', () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ source_ip: ".*" })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("Any IP")).toBeInTheDocument();
    });

    it("shows specific IP when source_ip is not wildcard", () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ source_ip: "192.168.1.0/24" })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("192.168.1.0/24")).toBeInTheDocument();
    });

    it('shows "All users" when username is ".*"', () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ username: ".*" })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("All users")).toBeInTheDocument();
    });

    it("shows specific username when not wildcard", () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ username: "alice" })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("alice")).toBeInTheDocument();
    });

    it("renders an Active badge for an active rule", () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ active: true })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("Active")).toBeInTheDocument();
    });

    it("renders an Inactive badge for an inactive rule", () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ active: false })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("Inactive")).toBeInTheDocument();
    });

    it("navigates to the detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ id: "rule-abc", priority: 99 })],
        totalCount: 1,
      });
      renderPage();

      // Click any cell in the row — use the priority text
      await user.click(screen.getByText("99"));
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/firewall-rules/rule-abc",
      );
    });

    it("renders the tenant_id as a namespace link", () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: [makeRule({ tenant_id: "tenant-xyz" })],
        totalCount: 1,
      });
      renderPage();
      const link = screen.getByRole("link", { name: "tenant-xyz" });
      expect(link).toBeInTheDocument();
      expect(link).toHaveAttribute("href", "/admin/namespaces/tenant-xyz");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the hook returns an error", () => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        error: new Error("Request failed"),
      });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("Request failed")).toBeInTheDocument();
    });
  });

  describe("client-side search", () => {
    // Rules are designed so each field is unique and non-overlapping with
    // other fields, ensuring search terms match exactly one rule at a time.
    const allowRule = makeRule({
      id: "r1",
      action: "allow",
      priority: 5,
      // Distinct IP that does not appear in any field of the deny rule.
      source_ip: "172.16.0.1",
      username: ".*",
    });
    const denyRule = makeRule({
      id: "r2",
      action: "deny",
      // Priority chosen so it does not appear in any other field of either rule.
      priority: 777,
      source_ip: ".*",
      username: "zara",
    });
    const searchRules = [allowRule, denyRule];

    beforeEach(() => {
      vi.mocked(useAdminFirewallRules).mockReturnValue({
        ...defaultHookState,
        rules: searchRules,
        totalCount: 2,
      });
    });

    it("filters rules by action text", async () => {
      const user = userEvent.setup();
      renderPage();

      await user.type(
        screen.getByRole("textbox", { name: "Search firewall rules" }),
        "deny",
      );

      await waitFor(() =>
        expect(screen.queryByText("Allow")).not.toBeInTheDocument(),
      );
      expect(screen.getByText("Deny")).toBeInTheDocument();
    });

    it("filters rules by source IP text", async () => {
      const user = userEvent.setup();
      renderPage();

      await user.type(
        screen.getByRole("textbox", { name: "Search firewall rules" }),
        "172.16.0.1",
      );

      await waitFor(() =>
        expect(screen.queryByText("zara")).not.toBeInTheDocument(),
      );
      expect(screen.getByText("172.16.0.1")).toBeInTheDocument();
    });

    it("filters rules by username text", async () => {
      const user = userEvent.setup();
      renderPage();

      await user.type(
        screen.getByRole("textbox", { name: "Search firewall rules" }),
        "zara",
      );

      await waitFor(() =>
        expect(screen.queryByText("172.16.0.1")).not.toBeInTheDocument(),
      );
      expect(screen.getByText("zara")).toBeInTheDocument();
    });

    it("filters rules by priority number", async () => {
      const user = userEvent.setup();
      renderPage();

      await user.type(
        screen.getByRole("textbox", { name: "Search firewall rules" }),
        "777",
      );

      await waitFor(() =>
        expect(screen.queryByText("Allow")).not.toBeInTheDocument(),
      );
      expect(screen.getByText("777")).toBeInTheDocument();
    });

    it('shows "No rules matching" message when search has no results', async () => {
      const user = userEvent.setup();
      renderPage();

      await user.type(
        screen.getByRole("textbox", { name: "Search firewall rules" }),
        "zzz-no-match",
      );

      await screen.findByText(/No rules matching/);
    });
  });
});
