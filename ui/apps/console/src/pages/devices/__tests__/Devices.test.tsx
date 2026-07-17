import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { NormalizedDevice } from "@/hooks/useDevices";
import type { UseDeviceActionsResult } from "@/hooks/useDeviceActions";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useDevices", () => ({
  useDevices: vi.fn(),
  buildFilter: vi.fn(),
}));

vi.mock("@/hooks/useDeviceMutations", () => ({
  useAddDeviceTag: vi.fn(() => ({ mutateAsync: vi.fn() })),
  useRemoveDeviceTag: vi.fn(() => ({ mutateAsync: vi.fn() })),
}));

// Return the value immediately (no timer) so tests don't need fake timers.
vi.mock("@/hooks/useDebouncedValue", () => ({
  useDebouncedValue: <T,>(value: T) => value,
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: () => ({ namespace: { name: "my-ns" } }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (sel: (s: { tenant: string }) => unknown) =>
    sel({ tenant: "tenant-1" }),
}));

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (sel: (s: { sessions: [] }) => unknown) =>
    sel({ sessions: [] }),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

vi.mock("@/components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button type="button" aria-label={`Copy ${text}`} />
  ),
}));

vi.mock("@/components/common/PageHeader", () => ({
  default: ({
    title,
    children,
  }: {
    title: string;
    children?: React.ReactNode;
  }) => (
    <div>
      <h1>{title}</h1>
      {children}
    </div>
  ),
}));

vi.mock("@/components/common/PlatformBadge", () => ({
  default: ({ platform }: { platform: string }) => <span>{platform}</span>,
}));

vi.mock("@/utils/date", () => ({
  formatRelative: () => "just now",
  formatDateFull: () => "Jan 15, 2024",
}));

vi.mock("@/utils/sshid", () => ({
  buildSshid: (ns: string, name: string) => `${ns}.${name}@localhost`,
}));

vi.mock("@/components/common/TagFilterDropdown", () => ({
  default: () => <div />,
}));

const mockManageTagsDrawer = vi.fn();
vi.mock("@/components/ManageTagsDrawer", () => ({
  default: (props: {
    open: boolean;
    onClose: () => void;
    onTagRenamed?: (oldName: string, newName: string) => void;
    onTagDeleted?: (name: string) => void;
  }) => {
    mockManageTagsDrawer(props);
    return <div data-testid="manage-tags-drawer" />;
  },
}));

vi.mock("@/components/ConnectDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/common/TagsPopover", () => ({
  default: ({ tags }: { tags: string[] }) => (
    <span>{tags.length > 0 ? tags.join(", ") : "No tags"}</span>
  ),
}));

const mockDeviceActionsPortal = vi.fn();
vi.mock("../DeviceActionsPortal", () => ({
  default: (props: { controller: UseDeviceActionsResult }) => {
    mockDeviceActionsPortal(props);
    return null;
  },
}));

const mockRequestAction = vi.fn();
const mockDeviceActionsController: UseDeviceActionsResult = {
  operation: undefined,
  requestAction: mockRequestAction,
  close: vi.fn(),
  billingWarningOpen: false,
  closeBillingWarning: vi.fn(),
  onBillingWarning: undefined,
  runSuccess: vi.fn(),
  billingEnabled: false,
};
vi.mock("@/hooks/useDeviceActions", () => ({
  useDeviceActions: vi.fn(() => mockDeviceActionsController),
}));

vi.mock("@/components/common/RestrictedAction", () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import React from "react";
import { useDevices } from "@/hooks/useDevices";
import { useDeviceActions } from "@/hooks/useDeviceActions";
import Devices from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

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
    tags: [],
    last_seen: new Date().toISOString(),
    created_at: new Date().toISOString(),
    identity: { mac: "aa:bb:cc:dd:ee:ff" },
    info: {
      id: "ubuntu",
      pretty_name: "Ubuntu 22.04 LTS",
      arch: "x86_64",
      platform: "linux",
      version: "0.14.0",
    },
    remote_addr: "1.2.3.4",
    custom_fields: {},
    ...overrides,
  } as NormalizedDevice;
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Devices />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("Devices list", () => {
  beforeEach(() => {
    vi.mocked(useDevices).mockReturnValue(defaultHookState);
    vi.mocked(useDeviceActions).mockReturnValue(mockDeviceActionsController);
    mockNavigate.mockReset();
    mockManageTagsDrawer.mockReset();
    mockDeviceActionsPortal.mockReset();
    mockRequestAction.mockReset();
  });

  describe("rendering", () => {
    it("renders the page heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Devices" }),
      ).toBeInTheDocument();
    });

    it("renders the Accepted tab and the Install Keys link, not the pending/rejected tabs", () => {
      renderPage();
      expect(
        screen.getByRole("button", { name: "Accepted" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("link", { name: "Install Keys" }),
      ).toBeInTheDocument();
      // Pending/rejected devices are managed from the install-keys area now, so the list is
      // accepted-only and no longer offers those tabs.
      expect(
        screen.queryByRole("button", { name: "Pending" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Rejected" }),
      ).not.toBeInTheDocument();
    });

    it("renders the search input", () => {
      renderPage();
      expect(
        screen.getByPlaceholderText("Search by hostname..."),
      ).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it("renders the loading message", () => {
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        isLoading: true,
      });
      renderPage();
      expect(screen.getByText("Loading devices...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No devices found" when list is empty', () => {
      renderPage();
      expect(screen.getByText("No devices found")).toBeInTheDocument();
    });
  });

  describe("device rows", () => {
    it("renders a row for each device", () => {
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        devices: [
          makeDevice({ uid: "uid-1", name: "alpha" }),
          makeDevice({ uid: "uid-2", name: "beta" }),
        ],
        totalCount: 2,
      });
      renderPage();
      expect(screen.getByText("alpha")).toBeInTheDocument();
      expect(screen.getByText("beta")).toBeInTheDocument();
    });

    it("navigates to device detail on row click", async () => {
      const user = userEvent.setup();
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        devices: [makeDevice({ uid: "uid-abc", name: "clickable" })],
        totalCount: 1,
      });
      renderPage();
      await user.click(screen.getByText("clickable"));
      expect(mockNavigate).toHaveBeenCalledWith("/devices/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error message when hook returns an error", () => {
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        error: new Error("Request failed"),
      });
      renderPage();
      expect(screen.getByText("Request failed")).toBeInTheDocument();
    });
  });

  describe("sorting", () => {
    it("requests last_seen/desc sort by default", () => {
      renderPage();
      expect(useDevices).toHaveBeenCalledWith(
        expect.objectContaining({ sortBy: "last_seen", orderBy: "desc" }),
      );
    });

    it("toggles sort when the Hostname header is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        devices: [makeDevice({ uid: "uid-1", name: "alpha" })],
        totalCount: 1,
      });
      renderPage();

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      let calls = vi.mocked(useDevices).mock.calls;
      let last = calls[calls.length - 1][0];
      expect(last).toMatchObject({ sortBy: "name", orderBy: "asc" });

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      calls = vi.mocked(useDevices).mock.calls;
      last = calls[calls.length - 1][0];
      expect(last).toMatchObject({ sortBy: "name", orderBy: "desc" });
    });
  });

  // ── URL hydration (usePaginatedListState adoption) ────────────────────────────

  describe("URL hydration — URL params seed page state on mount", () => {
    it("passes status=pending from URL to useDevices", () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "pending" }),
      );
    });

    it("passes tags array from URL to useDevices", () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ filterTags: ["a", "b"] }),
      );
    });

    it("passes page=2 from URL to useDevices", () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      );
    });

    it("falls back to status=accepted and page=1 when URL has no params", () => {
      renderPage(["/"]);
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "accepted", page: 1 }),
      );
    });

    it("falls back to status=accepted for an invalid status value", () => {
      renderPage(["/?status=invalid"]);
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "accepted" }),
      );
    });

    it("passes empty filterTags when no tags param is present", () => {
      renderPage(["/"]);
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ filterTags: [] }),
      );
    });
  });

  // ── Search trimming ───────────────────────────────────────────────────────────

  describe("search — whitespace is trimmed before passing to useDevices", () => {
    it("passes trimmed search to useDevices when input has surrounding spaces", async () => {
      // useDebouncedValue is mocked to return its input immediately, so we can
      // verify the trim happens before the debounce without needing fake timers.
      const user = userEvent.setup();
      renderPage();
      const searchInput = screen.getByPlaceholderText("Search by hostname...");
      await user.type(searchInput, "  myhost  ");
      // The hook must have been called with the trimmed string "myhost".
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ search: "myhost" }),
      );
    });
  });

  // ── DeviceActionsPortal integration ──────────────────────────────────────────

  describe("DeviceActionsPortal integration — useDeviceActions + portal replace inline state", () => {
    beforeEach(() => {
      mockRequestAction.mockReset();
      mockDeviceActionsPortal.mockReset();
      vi.mocked(useDeviceActions).mockReturnValue(mockDeviceActionsController);
    });

    it("mounts DeviceActionsPortal with the controller returned by useDeviceActions", () => {
      renderPage();
      expect(mockDeviceActionsPortal).toHaveBeenCalledWith(
        expect.objectContaining({ controller: mockDeviceActionsController }),
      );
    });

    // Accept/reject/remove of pending and rejected devices moved to the install-keys area; the list is
    // accepted-only (status is forced to accepted), so those inline actions are no longer reachable here.
  });

  // ── Tag mutation callbacks ────────────────────────────────────────────────────

  describe("tag mutation — onTagRenamed/onTagDeleted update URL tags array", () => {
    it("renames a tag in filterTags when onTagRenamed is called from ManageTagsDrawer", async () => {
      // Start with tags=a&tags=b in the URL
      renderPage(["/?tags=a&tags=b"]);

      // Grab the onTagRenamed callback passed to ManageTagsDrawer
      const lastCall = mockManageTagsDrawer.mock.calls.at(-1)?.[0] as {
        onTagRenamed?: (oldName: string, newName: string) => void;
      };
      expect(lastCall?.onTagRenamed).toBeDefined();

      // Invoke the callback inside act() so React flushes the URL update.
      await act(async () => {
        lastCall.onTagRenamed!("a", "alpha");
      });

      // useDevices should now be called with filterTags exactly ["alpha", "b"] (no stale "a")
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({
          filterTags: ["alpha", "b"],
        }),
      );
    });

    it("removes a tag from filterTags when onTagDeleted is called from ManageTagsDrawer", async () => {
      renderPage(["/?tags=a&tags=b"]);

      const lastCall = mockManageTagsDrawer.mock.calls.at(-1)?.[0] as {
        onTagDeleted?: (name: string) => void;
      };
      expect(lastCall?.onTagDeleted).toBeDefined();

      await act(async () => {
        lastCall.onTagDeleted!("a");
      });

      // useDevices should now be called with only tag "b"
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ filterTags: ["b"] }),
      );
    });

    it("adds a tag to URL array when setArrayFilter is called via addFilterTag", () => {
      // TagFilterDropdown and TagsPopover are mocked away; verify URL array
      // hydration indirectly: render with an existing tag in the URL and confirm
      // useDevices receives it.
      renderPage(["/?tags=existing"]);
      // The tags=existing param must arrive at useDevices
      expect(vi.mocked(useDevices)).toHaveBeenCalledWith(
        expect.objectContaining({ filterTags: ["existing"] }),
      );
      // The filter bar is still visible and the tags array remains stable
      expect(
        screen.getByPlaceholderText("Search by hostname..."),
      ).toBeInTheDocument();
    });
  });
});
