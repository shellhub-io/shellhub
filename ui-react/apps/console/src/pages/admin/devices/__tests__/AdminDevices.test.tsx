import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { NormalizedDevice } from "@/hooks/useAdminDevices";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useAdminDevices", () => ({
  useAdminDevices: vi.fn(),
}));

// useNavigate is used by the page — mock at the module level.
const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// ── Helpers ───────────────────────────────────────────────────────────────────

import { useAdminDevices } from "@/hooks/useAdminDevices";
import AdminDevices from "../index";

const defaultHookState = {
  devices: [] as NormalizedDevice[],
  totalCount: 0,
  isLoading: false,
  error: null,
  refetch: vi.fn(),
};

function makeDevice(
  overrides: Partial<NormalizedDevice> = {},
): NormalizedDevice {
  return {
    uid: "device-uid-1",
    name: "my-device",
    status: "accepted",
    online: true,
    namespace: "my-namespace",
    tenant_id: "tenant-1",
    tags: [],
    last_seen: new Date().toISOString(),
    created_at: new Date().toISOString(),
    identity: { mac: "aa:bb:cc:dd:ee:ff" },
    info: {
      id: "ubuntu",
      pretty_name: "Ubuntu 22.04",
      arch: "x86_64",
      platform: "linux",
      version: "0.14.0",
    },
    remote_addr: "1.2.3.4",
    public_key: null,
    ...overrides,
  } as NormalizedDevice;
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminDevices />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AdminDevices", () => {
  beforeEach(() => {
    vi.mocked(useAdminDevices).mockReturnValue(defaultHookState);
    mockNavigate.mockReset();
  });

  describe("rendering", () => {
    it("renders the page heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Devices" }),
      ).toBeInTheDocument();
    });

    it("renders the search input with correct aria-label", () => {
      renderPage();
      expect(
        screen.getByRole("textbox", { name: "Search devices by hostname" }),
      ).toBeInTheDocument();
    });

    it("renders all status filter tabs", () => {
      renderPage();
      expect(screen.getByRole("tab", { name: "All" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Accepted" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Pending" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Rejected" })).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading devices..." text', () => {
      vi.mocked(useAdminDevices).mockReturnValue({
        ...defaultHookState,
        isLoading: true,
        devices: [],
      });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading devices...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No devices found" when the device list is empty', () => {
      renderPage();
      expect(screen.getByText("No devices found")).toBeInTheDocument();
    });
  });

  describe("device rows", () => {
    it("renders a row for each returned device", () => {
      vi.mocked(useAdminDevices).mockReturnValue({
        ...defaultHookState,
        devices: [
          makeDevice({ uid: "uid-1", name: "device-alpha" }),
          makeDevice({ uid: "uid-2", name: "device-beta" }),
        ],
        totalCount: 2,
      });
      renderPage();
      expect(screen.getByText("device-alpha")).toBeInTheDocument();
      expect(screen.getByText("device-beta")).toBeInTheDocument();
    });

    it("renders the status chip for each device", () => {
      vi.mocked(useAdminDevices).mockReturnValue({
        ...defaultHookState,
        devices: [makeDevice({ status: "pending" })],
        totalCount: 1,
      });
      renderPage();
      // "Pending" appears in both the filter tab button and the status chip span.
      // Assert that at least two elements carry the text — one tab + one chip.
      expect(screen.getAllByText("Pending").length).toBeGreaterThanOrEqual(2);
    });

    it("navigates to the device detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminDevices).mockReturnValue({
        ...defaultHookState,
        devices: [makeDevice({ uid: "uid-abc", name: "clickable-device" })],
        totalCount: 1,
      });
      renderPage();

      await user.click(screen.getByText("clickable-device"));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/devices/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the hook returns an error", () => {
      vi.mocked(useAdminDevices).mockReturnValue({
        ...defaultHookState,
        error: new Error("Request failed"),
      });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("Request failed")).toBeInTheDocument();
    });
  });

  describe("status tab interaction", () => {
    it("calls useAdminDevices — status tab click re-renders without crashing", async () => {
      const user = userEvent.setup();
      renderPage();
      await user.click(screen.getByRole("tab", { name: "Accepted" }));
      // After clicking a tab the hook is still called; the page should still render
      expect(
        screen.getByRole("textbox", { name: "Search devices by hostname" }),
      ).toBeInTheDocument();
    });
  });
});
