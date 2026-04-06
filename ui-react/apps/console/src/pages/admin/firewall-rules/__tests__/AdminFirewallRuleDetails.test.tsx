import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { FirewallRule } from "../../../../hooks/useAdminFirewallRules";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("../../../../hooks/useAdminFirewallRules", () => ({
  useAdminFirewallRule: vi.fn(),
}));

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useParams: () => ({ id: "rule-1" }) };
});

// CopyButton relies on ClipboardProvider context and calls showModal() via
// BaseDialog, which is not supported in jsdom. Mock it to a simple no-op.
vi.mock("@/components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button type="button" aria-label={`Copy ${text}`} />
  ),
}));

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { useAdminFirewallRule } from "../../../../hooks/useAdminFirewallRules";
import AdminFirewallRuleDetails from "../AdminFirewallRuleDetails";

// ── Helpers ───────────────────────────────────────────────────────────────────

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
      <AdminFirewallRuleDetails />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AdminFirewallRuleDetails", () => {
  beforeEach(() => {
    vi.mocked(useAdminFirewallRule).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useAdminFirewallRule>);
  });

  describe("loading state", () => {
    it('renders a loading spinner with sr-only "Loading firewall rule details" while loading', () => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      } as ReturnType<typeof useAdminFirewallRule>);

      renderPage();

      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(
        screen.getByText("Loading firewall rule details"),
      ).toBeInTheDocument();
    });
  });

  describe("not-found / error state", () => {
    it('renders "Firewall rule not found" when no data and no loading', () => {
      renderPage();
      expect(screen.getByText("Firewall rule not found")).toBeInTheDocument();
    });

    it('renders "Firewall rule not found" when the hook returns an error', () => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error("404 Not Found"),
      } as never);

      renderPage();
      expect(screen.getByText("Firewall rule not found")).toBeInTheDocument();
    });

    it('renders a "Back to firewall rules" link in the not-found state', () => {
      renderPage();
      expect(
        screen.getByRole("link", { name: "Back to firewall rules" }),
      ).toBeInTheDocument();
    });
  });

  describe("rule data — allow rule", () => {
    beforeEach(() => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: makeRule(),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminFirewallRule>);
    });

    it('renders "Allow Rule" as the main heading', () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Allow Rule" }),
      ).toBeInTheDocument();
    });

    it("renders the breadcrumb navigation with Firewall Rules link", () => {
      renderPage();
      expect(
        screen.getByRole("navigation", { name: "Breadcrumb" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("link", { name: "Firewall Rules" }),
      ).toBeInTheDocument();
    });

    it("renders the rule ID", () => {
      renderPage();
      expect(screen.getAllByText("rule-1").length).toBeGreaterThanOrEqual(1);
    });

    it("renders the namespace as a link to the admin namespace page", () => {
      renderPage();
      const link = screen.getByRole("link", { name: "tenant-abc" });
      expect(link).toBeInTheDocument();
      expect(link).toHaveAttribute("href", "/admin/namespaces/tenant-abc");
    });

    it("renders the priority number", () => {
      renderPage();
      // Priority appears as both "Priority 1" badge and as the value in the card
      expect(screen.getByText("1")).toBeInTheDocument();
    });

    it('renders "Allow" in the action field', () => {
      renderPage();
      // The action label appears at least in the connection criteria card
      expect(screen.getAllByText("Allow").length).toBeGreaterThanOrEqual(1);
    });

    it("renders the Active badge", () => {
      renderPage();
      // "Active" appears in both the header badge and the properties card.
      expect(screen.getAllByText("Active").length).toBeGreaterThanOrEqual(1);
    });

    it('renders "Any IP" when source_ip is ".*"', () => {
      renderPage();
      expect(screen.getByText("Any IP")).toBeInTheDocument();
    });

    it('renders "All users" when username is ".*"', () => {
      renderPage();
      expect(screen.getByText("All users")).toBeInTheDocument();
    });

    it('renders "All devices" FilterBadge when filter hostname is ".*"', () => {
      renderPage();
      expect(screen.getByText("All devices")).toBeInTheDocument();
    });
  });

  describe("rule data — deny rule", () => {
    it('renders "Deny Rule" as the main heading', () => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: makeRule({ action: "deny" }),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminFirewallRule>);

      renderPage();
      expect(
        screen.getByRole("heading", { name: "Deny Rule" }),
      ).toBeInTheDocument();
    });
  });

  describe("rule data — inactive rule", () => {
    it("renders the Inactive badge", () => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: makeRule({ active: false }),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminFirewallRule>);

      renderPage();
      // The component renders "Inactive" in both the header badge and the
      // properties card — assert at least one is present.
      expect(screen.getAllByText("Inactive").length).toBeGreaterThanOrEqual(1);
    });
  });

  describe("rule data — specific IP and username", () => {
    it("renders a specific source IP when not wildcard", () => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: makeRule({ source_ip: "10.0.0.5" }),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminFirewallRule>);

      renderPage();
      expect(screen.getByText("10.0.0.5")).toBeInTheDocument();
    });

    it("renders a specific username when not wildcard", () => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: makeRule({ username: "alice" }),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminFirewallRule>);

      renderPage();
      expect(screen.getByText("alice")).toBeInTheDocument();
    });
  });

  describe("rule data — device filter", () => {
    it("renders hostname FilterBadge when filter has a specific hostname", () => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: makeRule({ filter: { hostname: "my-server" } }),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminFirewallRule>);

      renderPage();
      expect(screen.getByText("my-server")).toBeInTheDocument();
    });

    it("renders tag FilterBadge when filter has tags", () => {
      vi.mocked(useAdminFirewallRule).mockReturnValue({
        data: makeRule({ filter: { tags: ["production", "web"] } }),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminFirewallRule>);

      renderPage();
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("web")).toBeInTheDocument();
    });
  });
});
