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
    it('renders "Firewall rule not found" when no data and no loading', async () => {
      server.use(
        http.get("*/admin/api/firewall/rules/:id", () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      );
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Firewall rule not found")).toBeInTheDocument();
      });
    });

    it('renders "Firewall rule not found" when the query returns an error', async () => {
      server.use(
        http.get("*/admin/api/firewall/rules/:id", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Firewall rule not found")).toBeInTheDocument();
      });
    });

    it('renders a "Back to firewall rules" link in the not-found state', async () => {
      server.use(
        http.get("*/admin/api/firewall/rules/:id", () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      );
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("link", { name: "Back to firewall rules" }),
        ).toBeInTheDocument();
      });
    });
  });

  describe("rule data — allow rule", () => {
    it('renders "Allow Rule" as the main heading', async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: "Allow Rule" }),
        ).toBeInTheDocument();
      });
    });

    it("renders the rule ID", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getAllByText("rule-1").length).toBeGreaterThanOrEqual(1);
      });
    });

    it("renders the namespace as a link to the admin namespace page", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("link", { name: "tenant-abc" }),
        ).toBeInTheDocument();
      });
      expect(screen.getByRole("link", { name: "tenant-abc" })).toHaveAttribute(
        "href",
        "/admin/namespaces/tenant-abc",
      );
    });

    it("renders the priority number", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
      });
    });

    it('renders "Allow" in the action field', async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getAllByText("Allow").length).toBeGreaterThanOrEqual(1);
      });
    });

    it("renders the Active badge", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getAllByText("Active").length).toBeGreaterThanOrEqual(1);
      });
    });

    it('renders "Any IP" when source_ip is ".*"', async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Any IP")).toBeInTheDocument();
      });
    });

    it('renders "All users" when username is ".*"', async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("All users")).toBeInTheDocument();
      });
    });

    it('renders "All devices" FilterBadge when filter hostname is ".*"', async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("All devices")).toBeInTheDocument();
      });
    });
  });

  describe("rule data — deny rule", () => {
    it('renders "Deny Rule" as the main heading', async () => {
      setRule({ action: "deny" });
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: "Deny Rule" }),
        ).toBeInTheDocument();
      });
    });
  });

  describe("rule data — inactive rule", () => {
    it("renders the Inactive badge", async () => {
      setRule({ active: false });
      renderPage();
      await waitFor(() => {
        expect(screen.getAllByText("Inactive").length).toBeGreaterThanOrEqual(
          1,
        );
      });
    });
  });

  describe("rule data — specific IP and username", () => {
    it("renders a specific source IP when not wildcard", async () => {
      setRule({ source_ip: "10.0.0.5" });
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("10.0.0.5")).toBeInTheDocument();
      });
    });

    it("renders a specific username when not wildcard", async () => {
      setRule({ username: "alice" });
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("alice")).toBeInTheDocument();
      });
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
