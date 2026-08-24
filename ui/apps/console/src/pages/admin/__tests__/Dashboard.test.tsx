import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import AdminDashboard from "../Dashboard";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getStats: vi.fn(),
  }),
);

vi.mock("@/components/sessions/RecentSessionsTable", () => ({
  default: ({ isAdmin }: { isAdmin?: boolean }) => (
    <div data-testid="recent-sessions-table" data-admin={isAdmin} />
  ),
}));

const fullStats = {
  registered_users: 42,
  registered_devices: 150,
  online_devices: 75,
  active_sessions: 12,
  pending_devices: 5,
  rejected_devices: 3,
};

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminDashboard />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  sdk.getStats.mockResolvedValue(mockSdkResponse(fullStats));
});

describe("AdminDashboard", () => {
  describe("loading state", () => {
    it("renders spinner with role='status'", () => {
      sdk.getStats.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("does not render page header while loading", () => {
      sdk.getStats.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.queryByText("System Overview")).not.toBeInTheDocument();
    });

    it("does not render stat cards while loading", () => {
      sdk.getStats.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.queryByText("Registered Users")).not.toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders error message with role='alert'", async () => {
      sdk.getStats.mockRejectedValue(makeSdkError(500));
      renderPage();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("displays the expected error message", async () => {
      sdk.getStats.mockRejectedValue(makeSdkError(500));
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByText("Failed to load dashboard statistics"),
        ).toBeInTheDocument();
      });
    });

    it("does not render stat cards on stats error", async () => {
      sdk.getStats.mockRejectedValue(makeSdkError(500));
      renderPage();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(screen.queryByText("Registered Users")).not.toBeInTheDocument();
    });

    it("does not render sessions table on stats error", async () => {
      sdk.getStats.mockRejectedValue(makeSdkError(500));
      renderPage();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(
        screen.queryByTestId("recent-sessions-table"),
      ).not.toBeInTheDocument();
    });
  });

  describe("success state — all fields present, sessions present", () => {
    it("renders page header with correct title", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("System Overview")).toBeInTheDocument();
      });
    });

    it("renders page header with correct overline", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Admin Dashboard")).toBeInTheDocument();
      });
    });

    it("renders all six stat card titles", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Registered Users")).toBeInTheDocument();
      });
      expect(screen.getByText("Registered Devices")).toBeInTheDocument();
      expect(screen.getByText("Online Devices")).toBeInTheDocument();
      expect(screen.getByText("Active Sessions")).toBeInTheDocument();
      expect(screen.getByText("Pending Devices")).toBeInTheDocument();
      expect(screen.getByText("Rejected Devices")).toBeInTheDocument();
    });

    it("renders correct numeric values for each stat", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("42")).toBeInTheDocument();
      });
      expect(screen.getByText("150")).toBeInTheDocument();
      expect(screen.getByText("75")).toBeInTheDocument();
      expect(screen.getByText("12")).toBeInTheDocument();
      expect(screen.getByText("5")).toBeInTheDocument();
      expect(screen.getByText("3")).toBeInTheDocument();
    });

    it("'View all Users' link points to /admin/users", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("link", { name: /view all users/i }),
        ).toBeInTheDocument();
      });
      expect(
        screen.getByRole("link", { name: /view all users/i }),
      ).toHaveAttribute("href", "/admin/users");
    });

    it("'View all Sessions' link in stat card points to /admin/sessions", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("link", { name: /view all sessions/i }),
        ).toBeInTheDocument();
      });
      expect(
        screen.getByRole("link", { name: /view all sessions/i }),
      ).toHaveAttribute("href", "/admin/sessions");
    });

    it("device card links point to /admin/devices", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Registered Devices")).toBeInTheDocument();
      });
      const deviceLinks = screen.getAllByRole("link", { name: /devices/i });
      deviceLinks.forEach((link) => {
        expect(link).toHaveAttribute("href", "/admin/devices");
      });
    });

    it("renders RecentSessionsTable with isAdmin", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByTestId("recent-sessions-table")).toBeInTheDocument();
      });
      expect(screen.getByTestId("recent-sessions-table")).toHaveAttribute(
        "data-admin",
        "true",
      );
    });
  });

  describe("success state — partial stats response", () => {
    it("renders 0 for each missing stat field", async () => {
      sdk.getStats.mockResolvedValue(mockSdkResponse({ registered_users: 10 }));
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("10")).toBeInTheDocument();
      });
      const zeros = screen.getAllByText("0");
      expect(zeros.length).toBeGreaterThanOrEqual(5);
    });

    it("renders all zeros when stats is an empty object", async () => {
      sdk.getStats.mockResolvedValue(mockSdkResponse({}));
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("System Overview")).toBeInTheDocument();
      });
      const zeros = screen.getAllByText("0");
      expect(zeros.length).toBeGreaterThanOrEqual(6);
    });
  });
});
