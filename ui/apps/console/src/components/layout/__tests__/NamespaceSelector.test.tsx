import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { defaultConfig } from "@/env";
import NamespaceSelector from "../NamespaceSelector";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const mockGetConfig = vi.fn();

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: (): unknown => mockGetConfig() };
});

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespaces: () => ({
    namespaces: [],
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
  useNamespace: () => ({
    namespace: null,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
  useInitRole: () => {},
}));

vi.mock("@/hooks/useNamespaceMutations", () => ({
  useSwitchNamespace: () => ({
    mutateAsync: vi.fn(),
    isPending: false,
  }),
}));

vi.mock("@/components/common/CreateNamespaceDialog", () => ({
  default: () => null,
}));

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </MemoryRouter>
  );
}

function renderSelector(isAdminContext = false) {
  return render(<NamespaceSelector isAdminContext={isAdminContext} />, {
    wrapper: createWrapper(),
  });
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  useAuthStore.setState({
    token: "test-token",
    user: "alice",
    userId: "user-1",
    email: "alice@example.com",
    tenant: "t1",
    role: "owner",
    name: "Alice",
    isAdmin: false,
    loading: false,
  });
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("NamespaceSelector", () => {
  describe("showAdminLink", () => {
    // Case 1 — enterprise=true and cloud=true: link must appear and navigate to /admin.
    it("shows Admin Console link when enterprise=true, cloud=true, isAdmin=true, isAdminContext=false", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: true,
      });
      useAuthStore.setState({ tenant: "t1", isAdmin: true });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      const adminLink = screen.getByRole("button", { name: /admin console/i });
      expect(adminLink).toBeInTheDocument();
      await userEvent.click(adminLink);
      expect(mockNavigate).toHaveBeenCalledWith("/admin");
    });

    // Case 2 — already passes with current logic (enterprise=true, cloud=false).
    it("shows Admin Console link when enterprise=true, cloud=false, isAdmin=true, isAdminContext=false", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: false,
      });
      useAuthStore.setState({ tenant: "t1", isAdmin: true });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      const adminLink = screen.getByRole("button", { name: /admin console/i });
      expect(adminLink).toBeInTheDocument();
      await userEvent.click(adminLink);
      expect(mockNavigate).toHaveBeenCalledWith("/admin");
    });

    // Case 7 — cloud-only admin: the key case this fix enables.
    // cloud=true, enterprise=false → admin link must appear and navigate to /admin.
    it("shows Admin Console link when cloud=true, enterprise=false, isAdmin=true, isAdminContext=false", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: false,
        cloud: true,
      });
      useAuthStore.setState({ tenant: "t1", isAdmin: true });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      const adminLink = screen.getByRole("button", { name: /admin console/i });
      expect(adminLink).toBeInTheDocument();
      await userEvent.click(adminLink);
      expect(mockNavigate).toHaveBeenCalledWith("/admin");
    });

    // Case 3 — community instance (no enterprise, no cloud): link must be absent.
    it("hides Admin Console link when enterprise=false, cloud=false, isAdmin=true, isAdminContext=false", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: false,
        cloud: false,
      });
      useAuthStore.setState({ tenant: "t1", isAdmin: true });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      // Confirm the dropdown opened before asserting the link is absent.
      expect(screen.getByText("No namespaces available")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /admin console/i }),
      ).not.toBeInTheDocument();
    });

    // Case 4 — non-admin user on enterprise: link must be absent.
    it("hides Admin Console link when enterprise=true, cloud=false, isAdmin=false, isAdminContext=false", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: false,
      });
      useAuthStore.setState({ tenant: "t1", isAdmin: false });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      // Confirm the dropdown opened before asserting the link is absent.
      expect(screen.getByText("No namespaces available")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /admin console/i }),
      ).not.toBeInTheDocument();
    });

    // Case 5 — already inside admin context: showAdminLink must be false
    // because !isAdminContext short-circuits before enterprise/cloud.
    // The trigger is labeled "Admin Console" when isAdminContext=true; the dropdown
    // shows an "Admin Console" header (a <p>), but NOT the footer link button.
    it("hides Admin Console link when enterprise=true, cloud=true, isAdmin=true, isAdminContext=true", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: true,
      });
      useAuthStore.setState({ tenant: "t1", isAdmin: true });
      renderSelector(true);
      await userEvent.click(
        screen.getByRole("button", { name: /^admin console$/i }),
      );
      // Confirm the dropdown opened — the admin context header subtitle is visible.
      expect(screen.getByText("Super Admin · Instance")).toBeInTheDocument();
      // The footer link button (accessible name includes "Super Admin") must not render
      // in admin context — !isAdminContext suppresses showAdminLink.
      expect(
        screen.queryByRole("button", { name: /super admin/i }),
      ).not.toBeInTheDocument();
    });

    // Case 6 — cloud=true but non-admin user: isAdmin=false means link must be absent.
    it("hides Admin Console link when enterprise=false, cloud=true, isAdmin=false, isAdminContext=false", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: false,
        cloud: true,
      });
      useAuthStore.setState({ tenant: "t1", isAdmin: false });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      // Confirm the dropdown opened before asserting the link is absent.
      expect(screen.getByText("No namespaces available")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /admin console/i }),
      ).not.toBeInTheDocument();
    });
  });
});
