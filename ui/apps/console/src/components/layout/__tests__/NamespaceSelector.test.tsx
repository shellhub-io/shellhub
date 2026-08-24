import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { paginatedResponse, mockSdkResponse } from "@/tests/sdk";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { defaultConfig, getConfig } from "@/env";
import NamespaceSelector from "../NamespaceSelector";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getNamespaces: vi.fn(),
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
  }),
);

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/components/common/CreateNamespaceDialog", () => ({
  default: () => null,
}));

const mockGetConfig = vi.mocked(getConfig);

function renderSelector(isAdminContext = false) {
  return render(<NamespaceSelector isAdminContext={isAdminContext} />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  seedAuthStore();
  sdk.getNamespaces.mockResolvedValue(paginatedResponse([]));
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(null));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
});

describe("NamespaceSelector", () => {
  describe("showAdminLink", () => {
    it("shows Admin Console link in a cloud instance for admins not in the Admin context", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      seedAuthStore({ isAdmin: true });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      const adminLink = screen.getByRole("button", { name: /admin console/i });
      expect(adminLink).toBeInTheDocument();
      await userEvent.click(adminLink);
      expect(mockNavigate).toHaveBeenCalledWith("/admin");
    });

    it("shows Admin Console link in an Enterprise instance for admins not in the Admin context", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      seedAuthStore({ isAdmin: true });
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
      seedAuthStore({ isAdmin: true });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      expect(screen.getByText("No namespaces available")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /admin console/i }),
      ).not.toBeInTheDocument();
    });

    it("hides Admin Console link in an Enterprise instance for non-admin users", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      seedAuthStore({ isAdmin: false });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      expect(screen.getByText("No namespaces available")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /admin console/i }),
      ).not.toBeInTheDocument();
    });

    it("hides Admin Console link in the Admin context", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      seedAuthStore({ isAdmin: true });
      renderSelector(true);
      await userEvent.click(
        screen.getByRole("button", { name: /^admin console$/i }),
      );
      expect(screen.getByText("Super Admin · Instance")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /super admin/i }),
      ).not.toBeInTheDocument();
    });

    it("hides Admin Console link for non-admin users in a cloud instance", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      seedAuthStore({ isAdmin: false });
      renderSelector(false);
      await userEvent.click(
        screen.getByRole("button", { name: /select namespace/i }),
      );
      expect(screen.getByText("No namespaces available")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /admin console/i }),
      ).not.toBeInTheDocument();
    });
  });
});
