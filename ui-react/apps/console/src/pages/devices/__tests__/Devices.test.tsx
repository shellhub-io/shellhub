import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { NormalizedDevice } from "@/hooks/useDevices";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useDevices", () => ({
  useDevices: vi.fn(),
  buildFilter: vi.fn(),
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
  default: ({ title, children }: { title: string; children?: React.ReactNode }) => (
    <div><h1>{title}</h1>{children}</div>
  ),
}));

vi.mock("@/components/common/DataTable", () => ({
  default: ({
    columns,
    data,
    isLoading,
    loadingMessage,
    onRowClick,
    emptyState,
  }: {
    columns: { key: string; header: string; render: (row: unknown) => React.ReactNode }[];
    data: unknown[];
    isLoading: boolean;
    loadingMessage: string;
    onRowClick: (row: unknown) => void;
    emptyState: React.ReactNode;
  }) => {
    if (isLoading) return <div>{loadingMessage}</div>;
    if (data.length === 0) return <div>{emptyState}</div>;
    return (
      <table>
        <thead>
          <tr>{columns.map((c) => <th key={c.key}>{c.header}</th>)}</tr>
        </thead>
        <tbody>
          {data.map((row, i) => (
            <tr key={i} onClick={() => onRowClick(row)}>
              {columns.map((c) => <td key={c.key}>{c.render(row)}</td>)}
            </tr>
          ))}
        </tbody>
      </table>
    );
  },
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

vi.mock("@/components/ManageTagsDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/ConnectDrawer", () => ({
  default: () => <div />,
}));

vi.mock("../TagsPopover", () => ({
  default: ({ device }: { device: NormalizedDevice }) => (
    <span>{device.tags.length > 0 ? device.tags.join(", ") : "No tags"}</span>
  ),
}));

vi.mock("../DeviceActionDialog", () => ({
  default: () => <div />,
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
import Devices from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

const defaultHookState = {
  devices: [] as NormalizedDevice[],
  totalCount: 0,
  isLoading: false,
  error: null,
  refetch: vi.fn(),
};

function makeDevice(overrides: Partial<NormalizedDevice> = {}): NormalizedDevice {
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

function renderPage() {
  return render(
    <MemoryRouter>
      <Devices />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("Devices list", () => {
  beforeEach(() => {
    vi.mocked(useDevices).mockReturnValue(defaultHookState);
    mockNavigate.mockReset();
  });

  describe("rendering", () => {
    it("renders the page heading", () => {
      renderPage();
      expect(screen.getByRole("heading", { name: "Devices" })).toBeInTheDocument();
    });

    it("renders all status filter tabs", () => {
      renderPage();
      expect(screen.getByRole("button", { name: "Accepted" })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Pending" })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Rejected" })).toBeInTheDocument();
    });

    it("renders the search input", () => {
      renderPage();
      expect(screen.getByPlaceholderText("Search devices...")).toBeInTheDocument();
    });

    it('renders the "Custom Fields" column header', () => {
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        devices: [makeDevice()],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByRole("columnheader", { name: "Custom Fields" })).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it("renders the loading message", () => {
      vi.mocked(useDevices).mockReturnValue({ ...defaultHookState, isLoading: true });
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

  describe("custom fields column", () => {
    it("renders '—' when device has no custom fields", () => {
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        devices: [makeDevice({ custom_fields: {} })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("—")).toBeInTheDocument();
    });

    it("renders key and value badges for each custom field", () => {
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        devices: [makeDevice({ custom_fields: { env: "production", owner: "team-a" } })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("env")).toBeInTheDocument();
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("owner")).toBeInTheDocument();
      expect(screen.getByText("team-a")).toBeInTheDocument();
    });

    it("renders multiple custom field badges for multiple fields", () => {
      vi.mocked(useDevices).mockReturnValue({
        ...defaultHookState,
        devices: [makeDevice({ custom_fields: { a: "1", b: "2", c: "3" } })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("a")).toBeInTheDocument();
      expect(screen.getByText("1")).toBeInTheDocument();
      expect(screen.getByText("b")).toBeInTheDocument();
      expect(screen.getByText("2")).toBeInTheDocument();
      expect(screen.getByText("c")).toBeInTheDocument();
      expect(screen.getByText("3")).toBeInTheDocument();
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
});
