import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import AdminDashboard from "../Dashboard";
import { useAdminStats } from "@/hooks/useAdminStats";

vi.mock("@/hooks/useAdminStats", () => ({
  useAdminStats: vi.fn(),
}));

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

function setupHooks({
  statsData = fullStats,
  statsLoading = false,
  statsError = false,
} = {}) {
  vi.mocked(useAdminStats).mockReturnValue({
    stats: statsData,
    isLoading: statsLoading,
    isError: statsError,
  } as never);
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminDashboard />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("AdminDashboard", () => {
  describe("loading state", () => {
    it("renders spinner with role='status'", () => {
      setupHooks({ statsLoading: true });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("does not render page header while loading", () => {
      setupHooks({ statsLoading: true });
      renderPage();
      expect(screen.queryByText("System Overview")).not.toBeInTheDocument();
    });

    it("does not render stat cards while loading", () => {
      setupHooks({ statsLoading: true });
      renderPage();
      expect(screen.queryByText("Registered Users")).not.toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders error message with role='alert'", () => {
      setupHooks({ statsError: true });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });

    it("displays the expected error message", () => {
      setupHooks({ statsError: true });
      renderPage();
      expect(
        screen.getByText("Failed to load dashboard statistics"),
      ).toBeInTheDocument();
    });

    it("does not render stat cards on stats error", () => {
      setupHooks({ statsError: true });
      renderPage();
      expect(screen.queryByText("Registered Users")).not.toBeInTheDocument();
    });

    it("does not render sessions table on stats error", () => {
      setupHooks({ statsError: true });
      renderPage();
      expect(
        screen.queryByTestId("recent-sessions-table"),
      ).not.toBeInTheDocument();
    });
  });

  describe("success state — all fields present, sessions present", () => {
    it("renders page header with correct title", () => {
      setupHooks();
      renderPage();
      expect(screen.getByText("System Overview")).toBeInTheDocument();
    });

    it("renders page header with correct overline", () => {
      setupHooks();
      renderPage();
      expect(screen.getByText("Admin Dashboard")).toBeInTheDocument();
    });

    it("renders all six stat card titles", () => {
      setupHooks();
      renderPage();
      expect(screen.getByText("Registered Users")).toBeInTheDocument();
      expect(screen.getByText("Registered Devices")).toBeInTheDocument();
      expect(screen.getByText("Online Devices")).toBeInTheDocument();
      expect(screen.getByText("Active Sessions")).toBeInTheDocument();
      expect(screen.getByText("Pending Devices")).toBeInTheDocument();
      expect(screen.getByText("Rejected Devices")).toBeInTheDocument();
    });

    it("renders correct numeric values for each stat", () => {
      setupHooks();
      renderPage();
      expect(screen.getByText("42")).toBeInTheDocument();
      expect(screen.getByText("150")).toBeInTheDocument();
      expect(screen.getByText("75")).toBeInTheDocument();
      expect(screen.getByText("12")).toBeInTheDocument();
      expect(screen.getByText("5")).toBeInTheDocument();
      expect(screen.getByText("3")).toBeInTheDocument();
    });

    it("'View all Users' link points to /admin/users", () => {
      setupHooks();
      renderPage();
      expect(
        screen.getByRole("link", { name: /view all users/i }),
      ).toHaveAttribute("href", "/admin/users");
    });

    it("'View all Sessions' link in stat card points to /admin/sessions", () => {
      setupHooks();
      renderPage();
      expect(
        screen.getByRole("link", { name: /view all sessions/i }),
      ).toHaveAttribute("href", "/admin/sessions");
    });

    it("device card links point to /admin/devices", () => {
      setupHooks();
      renderPage();
      const deviceLinks = screen.getAllByRole("link", { name: /devices/i });
      deviceLinks.forEach((link) => {
        expect(link).toHaveAttribute("href", "/admin/devices");
      });
    });

    it("renders RecentSessionsTable with isAdmin", () => {
      setupHooks();
      renderPage();
      const table = screen.getByTestId("recent-sessions-table");
      expect(table).toBeInTheDocument();
      expect(table).toHaveAttribute("data-admin", "true");
    });
  });

  describe("success state — partial stats response", () => {
    it("renders 0 for each missing stat field", () => {
      setupHooks({ statsData: { registered_users: 10 } as typeof fullStats });
      renderPage();
      expect(screen.getByText("10")).toBeInTheDocument();
      const zeros = screen.getAllByText("0");
      expect(zeros.length).toBeGreaterThanOrEqual(5);
    });

    it("renders all zeros when stats is an empty object", () => {
      setupHooks({ statsData: {} as typeof fullStats });
      renderPage();
      const zeros = screen.getAllByText("0");
      expect(zeros.length).toBeGreaterThanOrEqual(6);
    });
  });
});
