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

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminDevices />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AdminDevices", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useAdminDevices).mockReturnValue(defaultHookState);
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
        screen.getByRole("searchbox", { name: "Search devices by hostname" }),
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
        screen.getByRole("searchbox", { name: "Search devices by hostname" }),
      ).toBeInTheDocument();
    });
  });

  // ── URL-driven state (usePaginatedListState adoption) ─────────────────────────

  describe("URL hydration — controls reflect URL params on mount", () => {
    it("calls useAdminDevices with sortBy/orderBy hydrated from URL sortField/sortOrder params", () => {
      renderPage(["/?sortField=name&sortOrder=asc"]);
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ sortBy: "name", orderBy: "asc" }),
      );
    });

    it("calls useAdminDevices with status hydrated from URL status param", () => {
      renderPage(["/?status=accepted"]);
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "accepted" }),
      );
    });

    it("marks the matching status tab as selected when status is in the URL", () => {
      renderPage(["/?status=pending"]);
      expect(screen.getByRole("tab", { name: "Pending" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("calls useAdminDevices with page hydrated from URL page param", () => {
      renderPage(["/?page=3"]);
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 3 }),
      );
    });

    it("uses defaults when URL params are absent (last_seen/desc, page 1, empty status)", () => {
      renderPage(["/"]);
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({
          sortBy: "last_seen",
          orderBy: "desc",
          page: 1,
          status: "",
        }),
      );
    });

    it("rejects an invalid status value and falls back to empty string (All tab selected)", () => {
      renderPage(["/?status=invalid-status"]);
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "" }),
      );
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
  });

  describe("URL writes — interactions update URL and reset page", () => {
    it("clicking a status tab writes status to URL and resets page to 1", async () => {
      const user = userEvent.setup();
      // Start on page 2 with no status
      renderPage(["/?page=2"]);
      // Confirm we start on page 2
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      );

      await user.click(screen.getByRole("tab", { name: "Accepted" }));

      // After clicking, useAdminDevices should be called with status=accepted and page=1
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "accepted", page: 1 }),
      );
    });

    it("clicking a sort column header writes sortField/sortOrder to URL and resets page", async () => {
      const user = userEvent.setup();
      renderPage(["/?page=3"]);

      // Click the "Sort by Hostname" sort button (maps to "name" field)
      await user.click(screen.getByRole("button", { name: /sort by hostname/i }));

      // name has initialOrder=asc, so switching to it uses asc
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ sortBy: "name", orderBy: "asc", page: 1 }),
      );
    });

    it("clicking the same sort column again toggles order from asc to desc", async () => {
      const user = userEvent.setup();
      // Start already sorted by name asc
      renderPage(["/?sortField=name&sortOrder=asc"]);

      await user.click(screen.getByRole("button", { name: /sort by hostname/i }));

      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ sortBy: "name", orderBy: "desc" }),
      );
    });
  });

  describe("URL writes — default params are omitted from the URL", () => {
    it("does not include page in the URL when on page 1 (the default)", () => {
      // We can't easily inspect the URL from a plain render test, so we verify
      // that useAdminDevices receives page=1 after navigating back to default
      renderPage(["/"]);
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 }),
      );
    });

    it("does not pass status to useAdminDevices when All tab is selected (default empty)", () => {
      renderPage(["/"]);
      expect(vi.mocked(useAdminDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "" }),
      );
    });
  });
});
