import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import AdminDashboard from "../Dashboard";
import { useAdminStats } from "../../../hooks/useAdminStats";
import { useAdminSessions } from "../../../hooks/useAdminSessions";

vi.mock("../../../hooks/useAdminStats", () => ({
  useAdminStats: vi.fn(),
}));
vi.mock("../../../hooks/useAdminSessions", () => ({
  useAdminSessions: vi.fn(),
}));

const fullStats = {
  registered_users: 42,
  registered_devices: 150,
  online_devices: 75,
  active_sessions: 12,
  pending_devices: 5,
  rejected_devices: 3,
};

const mockSession = {
  uid: "session-001",
  active: true,
  authenticated: true,
  username: "root",
  device_uid: "device-001",
  device: {
    uid: "device-001",
    name: "web-server-01",
    online: true,
    info: { id: "ubuntu" },
  },
  events: { types: ["shell"] },
  started_at: new Date(Date.now() - 3 * 60 * 1000).toISOString(),
  tenant_id: "tenant-001",
  ip_address: "192.168.0.1",
  last_seen: new Date().toISOString(),
  recorded: false,
  type: "term" as const,
  term: "",
  position: {},
};

function setupHooks({
  statsData = fullStats as object | undefined,
  statsLoading = false,
  statsError = false,
  sessions = [mockSession] as object[],
  sessionsLoading = false,
  sessionsError = null as unknown,
} = {}) {
  vi.mocked(useAdminStats).mockReturnValue({
    stats: statsData,
    isLoading: statsLoading,
    isError: statsError,
  } as never);
  vi.mocked(useAdminSessions).mockReturnValue({
    sessions,
    totalCount: sessions.length,
    isLoading: sessionsLoading,
    error: sessionsError,
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
      expect(screen.getByText("Failed to load dashboard statistics")).toBeInTheDocument();
    });

    it("does not render stat cards on stats error", () => {
      setupHooks({ statsError: true });
      renderPage();
      expect(screen.queryByText("Registered Users")).not.toBeInTheDocument();
    });

    it("does not render sessions section on stats error", () => {
      setupHooks({ statsError: true });
      renderPage();
      expect(screen.queryByText("Recent Sessions")).not.toBeInTheDocument();
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
      expect(screen.getByRole("link", { name: /view all users/i })).toHaveAttribute("href", "/admin/users");
    });

    it("'View all Sessions' link in stat card points to /admin/sessions", () => {
      setupHooks();
      renderPage();
      expect(screen.getByRole("link", { name: /view all sessions/i })).toHaveAttribute("href", "/admin/sessions");
    });

    it("device card links point to /admin/devices", () => {
      setupHooks();
      renderPage();
      const deviceLinks = screen.getAllByRole("link", { name: /devices/i });
      deviceLinks.forEach((link) => {
        expect(link).toHaveAttribute("href", "/admin/devices");
      });
    });

    it("renders the sessions table", () => {
      setupHooks();
      renderPage();
      expect(screen.getByRole("table")).toBeInTheDocument();
    });

    it("renders session username in the table", () => {
      setupHooks();
      renderPage();
      expect(screen.getByText("root")).toBeInTheDocument();
    });

    it("renders session device name in the table", () => {
      setupHooks();
      renderPage();
      expect(screen.getByText("web-server-01")).toBeInTheDocument();
    });

    it("renders session type badge", () => {
      setupHooks();
      renderPage();
      expect(screen.getByText("shell")).toBeInTheDocument();
    });

    it("'View all' sessions link in section header points to /admin/sessions", () => {
      setupHooks();
      renderPage();
      const viewAllLink = screen.getByRole("link", { name: /view all →/i });
      expect(viewAllLink).toHaveAttribute("href", "/admin/sessions");
    });

    it("DeviceChip in session row renders without a link", () => {
      setupHooks();
      renderPage();
      const chipText = screen.getByText("web-server-01");
      expect(chipText.closest("a")).toBeNull();
    });
  });

  describe("success state — partial stats response", () => {
    it("renders 0 for each missing stat field", () => {
      setupHooks({ statsData: { registered_users: 10 } });
      renderPage();
      expect(screen.getByText("10")).toBeInTheDocument();
      const zeros = screen.getAllByText("0");
      expect(zeros.length).toBeGreaterThanOrEqual(5);
    });

    it("renders all zeros when stats is an empty object", () => {
      setupHooks({ statsData: {} });
      renderPage();
      const zeros = screen.getAllByText("0");
      expect(zeros.length).toBeGreaterThanOrEqual(6);
    });
  });

  describe("success state — no sessions", () => {
    it("renders 'No recent sessions' empty state", () => {
      setupHooks({ sessions: [] });
      renderPage();
      expect(screen.getByText("No recent sessions")).toBeInTheDocument();
    });

    it("does not render the table when sessions are empty", () => {
      setupHooks({ sessions: [] });
      renderPage();
      expect(screen.queryByRole("table")).not.toBeInTheDocument();
    });

    it("still renders stat cards when sessions are empty", () => {
      setupHooks({ sessions: [] });
      renderPage();
      expect(screen.getByText("Registered Users")).toBeInTheDocument();
    });
  });

  describe("success state — sessions query error", () => {
    it("hides the sessions section when sessions query fails", () => {
      setupHooks({ sessionsError: new Error("network failure") });
      renderPage();
      expect(screen.queryByText("Recent Sessions")).not.toBeInTheDocument();
    });

    it("still renders stat cards when sessions fail", () => {
      setupHooks({ sessionsError: new Error("network failure") });
      renderPage();
      expect(screen.getByText("Registered Users")).toBeInTheDocument();
    });
  });

  describe("sessions loading state", () => {
    it("hides the sessions section while sessions are loading", () => {
      setupHooks({ sessions: [], sessionsLoading: true });
      renderPage();
      expect(screen.queryByText("Recent Sessions")).not.toBeInTheDocument();
      expect(screen.queryByText("No recent sessions")).not.toBeInTheDocument();
      expect(screen.queryByRole("table")).not.toBeInTheDocument();
    });

    it("still renders stat cards while sessions load", () => {
      setupHooks({ sessions: [], sessionsLoading: true });
      renderPage();
      expect(screen.getByText("Registered Users")).toBeInTheDocument();
    });
  });

  describe("unauthenticated session", () => {
    it("renders warning icon with title for suspicious sessions", () => {
      const suspiciousSession = { ...mockSession, authenticated: false };
      setupHooks({ sessions: [suspiciousSession] });
      renderPage();
      expect(screen.getByTitle("Not authenticated")).toBeInTheDocument();
    });
  });
});
