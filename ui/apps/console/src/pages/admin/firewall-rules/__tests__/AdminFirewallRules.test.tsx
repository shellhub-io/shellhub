import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import AdminFirewallRules from "../index";
import { makeSdkError, paginatedResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { mockFirewallRule } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getFirewallRulesAdmin: vi.fn(),
  }),
);

const capturedDataTableProps: Record<string, unknown>[] = [];
vi.mock("@/components/common/DataTable", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("@/components/common/DataTable")>();
  return {
    ...actual,
    default: (props: Record<string, unknown>) => {
      capturedDataTableProps.push({ ...props });
      return actual.default(
        props as unknown as Parameters<typeof actual.default>[0],
      );
    },
  };
});

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
    capturedDataTableProps.length = 0;
    useAuthStore.setState({ isAdmin: true });
    sdk.getFirewallRulesAdmin.mockResolvedValue(paginatedResponse([]));
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
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
      ).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading firewall rules..." text', () => {
      sdk.getFirewallRulesAdmin.mockReturnValue(new Promise(() => {}));
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
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse(
          [
            mockFirewallRule({ id: "r1", priority: 1 }),
            mockFirewallRule({ id: "r2", priority: 2 }),
          ],
          2,
        ),
      );
      renderPage();
      await waitFor(() => expect(screen.getAllByText("Allow").length).toBe(2));
      expect(screen.getAllByText("1")[0]).toBeInTheDocument();
      expect(screen.getAllByText("2")[0]).toBeInTheDocument();
    });

    it('shows "Allow" with accent-green for an allow rule', async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ action: "allow" })]),
      );
      renderPage();
      expect(await screen.findByText("Allow")).toBeInTheDocument();
    });

    it('shows "Deny" for a deny rule', async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ action: "deny" })]),
      );
      renderPage();
      expect(await screen.findByText("Deny")).toBeInTheDocument();
    });

    it('shows "Any IP" when source_ip is ".*"', async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ source_ip: ".*" })]),
      );
      renderPage();
      expect(await screen.findByText("Any IP")).toBeInTheDocument();
    });

    it("shows specific IP when source_ip is not wildcard", async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ source_ip: "192.168.1.0/24" })]),
      );
      renderPage();
      expect(await screen.findByText("192.168.1.0/24")).toBeInTheDocument();
    });

    it('shows "All users" when username is ".*"', async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ username: ".*" })]),
      );
      renderPage();
      expect(await screen.findByText("All users")).toBeInTheDocument();
    });

    it("shows specific username when not wildcard", async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ username: "alice" })]),
      );
      renderPage();
      expect(await screen.findByText("alice")).toBeInTheDocument();
    });

    it("renders an Active badge for an active rule", async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ active: true })]),
      );
      renderPage();
      expect(await screen.findByText("Active")).toBeInTheDocument();
    });

    it("renders an Inactive badge for an inactive rule", async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ active: false })]),
      );
      renderPage();
      expect(await screen.findByText("Inactive")).toBeInTheDocument();
    });

    it("navigates to the detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ id: "rule-abc", priority: 99 })]),
      );
      renderPage();

      await user.click(await screen.findByText("99"));
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/firewall-rules/rule-abc",
      );
    });

    it("renders the tenant_id as a namespace link", async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ tenant_id: "tenant-xyz" })]),
      );
      renderPage();
      const link = await screen.findByRole("link", { name: "tenant-xyz" });
      expect(link).toHaveAttribute("href", "/admin/namespaces/tenant-xyz");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the SDK returns an error", async () => {
      sdk.getFirewallRulesAdmin.mockRejectedValue(makeSdkError(500));
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
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([allowRule, denyRule]),
      );
    });

    it("filters rules by action text", async () => {
      const user = userEvent.setup();
      renderPage();

      await screen.findByText("Allow");

      await user.type(
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
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

      await screen.findByText("172.16.0.1");

      await user.type(
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
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

      await screen.findByText("zara");

      await user.type(
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
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

      await screen.findByText("777");

      await user.type(
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
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

      await screen.findByText("Allow");

      await user.type(
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
        "zzz-no-match",
      );

      await screen.findByText(/No rules matching/);
    });
  });

  describe("pagination suppressed while searching", () => {
    beforeEach(() => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse(
          [mockFirewallRule({ id: "r1", action: "allow", priority: 1 })],
          1,
        ),
      );
    });

    it("passes page/totalPages/onPageChange to DataTable when search is empty", async () => {
      renderPage();
      await screen.findByText("Allow");
      const last = capturedDataTableProps.at(-1);
      expect(last).toBeDefined();
      expect(last).toHaveProperty("page");
      expect(last).toHaveProperty("totalPages");
      expect(last).toHaveProperty("onPageChange");
    });

    it("omits page/totalPages/onPageChange from DataTable while search is non-empty", async () => {
      const user = userEvent.setup();
      renderPage();

      await screen.findByText("Allow");

      await user.type(
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
        "allow",
      );

      await waitFor(() => {
        const last = capturedDataTableProps.at(-1);
        expect(last).toBeDefined();
        expect(last).not.toHaveProperty("page");
        expect(last).not.toHaveProperty("totalPages");
        expect(last).not.toHaveProperty("onPageChange");
      });
    });
  });

  describe("URL round-trips", () => {
    it("hydrates search from URL on mount", async () => {
      sdk.getFirewallRulesAdmin.mockResolvedValue(
        paginatedResponse([mockFirewallRule({ id: "r1", action: "allow" })]),
      );
      renderPage(["/?search=allow"]);
      expect(
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
      ).toHaveValue("allow");
    });

    it("hydrates page from URL and passes it to the SDK", async () => {
      renderPage(["/?page=3"]);
      await screen.findByText("No firewall rules found");
      expect(sdk.getFirewallRulesAdmin).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 3 }),
        }),
      );
    });

    it("passes page=1 to the SDK when URL has no params", async () => {
      renderPage(["/"]);
      await screen.findByText("No firewall rules found");
      expect(sdk.getFirewallRulesAdmin).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 1 }),
        }),
      );
    });

    it("setSearch resets page to 1 in the URL", async () => {
      const user = userEvent.setup();
      renderPage(["/?page=3"]);

      await screen.findByText("No firewall rules found");

      expect(sdk.getFirewallRulesAdmin).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 3 }),
        }),
      );

      await user.type(
        screen.getByRole("searchbox", {
          name: "Search firewall rules by action, priority, IP, or username",
        }),
        "allow",
      );

      await waitFor(() => {
        const calls = sdk.getFirewallRulesAdmin.mock.calls;
        const lastCall = calls.at(-1)![0];
        expect(lastCall).toBeDefined();
        expect(lastCall?.query?.page).toBe(1);
      });
    });
  });
});
