import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import type { FirewallRulesResponse } from "@/client/model";
import AdminFirewallRuleDetails from "../AdminFirewallRuleDetails";

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useParams: () => ({ id: "rule-1" }) };
});

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

function makeTag(name: string) {
  return {
    name,
    tenant_id: "tenant-abc",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };
}

function makeRule(
  overrides: Partial<FirewallRulesResponse> = {},
): FirewallRulesResponse {
  return {
    id: "rule-1",
    tenant_id: "tenant-abc",
    priority: 1,
    action: "allow" as const,
    active: true,
    source_ip: ".*",
    username: ".*",
    filter: { hostname: ".*", tags: [] },
    ...overrides,
  };
}

function setRule(overrides: Partial<FirewallRulesResponse> = {}) {
  server.use(
    http.get("*/admin/api/firewall/rules/:id", () =>
      HttpResponse.json(makeRule(overrides)),
    ),
  );
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminFirewallRuleDetails />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  setRule();
});

describe("AdminFirewallRuleDetails", () => {
  describe("loading state", () => {
    it('announces "Loading firewall rule details" while loading', () => {
      server.use(
        http.get("*/admin/api/firewall/rules/:id", () => new Promise(() => {})),
      );
      renderPage();
      expect(
        screen.getByRole("status", { name: "Loading firewall rule details" }),
      ).toBeInTheDocument();
    });
  });

  describe("not-found / error state", () => {
    it('renders "Firewall rule not found" with a way back', async () => {
      server.use(
        http.get("*/admin/api/firewall/rules/:id", () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      );
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Firewall rule not found")).toBeInTheDocument();
      });
      expect(
        screen.getByRole("link", { name: "Back to firewall rules" }),
      ).toBeInTheDocument();
    });
  });

  describe("rule data — allow rule", () => {
    it("renders the rule's fields", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: "Allow Rule" }),
        ).toBeInTheDocument();
      });
      expect(screen.getAllByText("rule-1").length).toBeGreaterThanOrEqual(1);
      expect(screen.getByRole("link", { name: "tenant-abc" })).toHaveAttribute(
        "href",
        "/admin/namespaces/tenant-abc",
      );
      expect(screen.getByText("1")).toBeInTheDocument();
      expect(screen.getAllByText("Allow").length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText("Active").length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText("Any IP")).toBeInTheDocument();
      expect(screen.getByText("All users")).toBeInTheDocument();
      expect(screen.getByText("All devices")).toBeInTheDocument();
    });
  });

  describe("rule data — deny rule, inactive, specific IP and username", () => {
    it("renders the rule's own action, state, IP and username", async () => {
      setRule({
        action: "deny",
        active: false,
        source_ip: "10.0.0.5",
        username: "alice",
      });
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: "Deny Rule" }),
        ).toBeInTheDocument();
      });
      expect(screen.getAllByText("Inactive").length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText("10.0.0.5")).toBeInTheDocument();
      expect(screen.getByText("alice")).toBeInTheDocument();
    });
  });

  describe("rule data — device filter", () => {
    it("renders hostname FilterBadge when filter has a specific hostname", async () => {
      setRule({ filter: { hostname: "my-server", tags: [] } });
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("my-server")).toBeInTheDocument();
      });
    });

    it("renders tag FilterBadge when filter has tags", async () => {
      setRule({
        filter: { tags: [makeTag("production"), makeTag("web")] },
      });
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("production")).toBeInTheDocument();
      });
      expect(screen.getByText("web")).toBeInTheDocument();
    });
  });
});
