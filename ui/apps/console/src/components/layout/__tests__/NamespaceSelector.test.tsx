import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { defaultConfig, getConfig } from "@/env";
import NamespaceSelector from "../NamespaceSelector";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});
const mockGetConfig = vi.mocked(getConfig);

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
    it("shows Admin Console link in a cloud instance for admins not in the Admin context", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
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

    // Case 2 — edition=enterprise.
    it("shows Admin Console link in a Enterprise instance for admins not in the Admin context", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
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

    it("hides Admin Console link in a community instance", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
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

    it("hides Admin Console link in a Enterprise instance for non-admin users", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
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

    it("hides Admin Console link in the Admin context", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
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

    it("hides Admin Console link for non-admin users in a cloud instance", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
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
