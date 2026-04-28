import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import type { Device } from "@/client";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useDevice", () => ({
  useDevice: vi.fn(),
}));

const mockUpdateCustomFields = vi.fn();

vi.mock("@/hooks/useDeviceMutations", () => ({
  useRenameDevice: () => ({ mutateAsync: vi.fn() }),
  useAddDeviceTag: () => ({ mutateAsync: vi.fn() }),
  useRemoveDeviceTag: () => ({ mutateAsync: vi.fn() }),
  useRemoveDevice: () => ({ mutateAsync: vi.fn() }),
  useUpdateDeviceCustomFields: () => ({ mutateAsync: mockUpdateCustomFields }),
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: () => ({ namespace: { name: "my-ns" } }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (sel: (s: { tenant: string }) => unknown) =>
    sel({ tenant: "tenant-1" }),
}));

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (sel: (s: { sessions: []; restore: () => void }) => unknown) =>
    sel({ sessions: [], restore: vi.fn() }),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

vi.mock("@/components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button type="button" aria-label={`Copy ${text}`} />
  ),
}));

vi.mock("@/components/common/PlatformBadge", () => ({
  default: ({ platform }: { platform: string }) => <span>{platform}</span>,
}));

vi.mock("@/components/ConnectDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/common/RestrictedAction", () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

vi.mock("@/pages/devices/DeviceActionDialog", () => ({
  default: () => <div />,
}));

vi.mock("@/utils/date", () => ({
  formatRelative: () => "just now",
  formatDateFull: () => "Jan 15, 2024",
}));

vi.mock("@/utils/sshid", () => ({
  buildSshid: (ns: string, name: string) => `${ns}.${name}@localhost`,
}));

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return {
    ...actual,
    useParams: () => ({ uid: "test-uid" }),
    useNavigate: () => vi.fn(),
    useSearchParams: () => [new URLSearchParams(), vi.fn()],
  };
});

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { useDevice } from "@/hooks/useDevice";
import DeviceDetails from "../../DeviceDetails";

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeDevice(overrides: Partial<Device> = {}): Device {
  return {
    uid: "test-uid",
    name: "my-device",
    status: "accepted",
    online: true,
    tags: [],
    last_seen: "2024-01-15T10:00:00.000Z",
    created_at: "2023-06-01T08:00:00.000Z",
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
  } as Device;
}

function renderPage() {
  return render(
    <MemoryRouter>
      <DeviceDetails />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("DeviceDetails", () => {
  beforeEach(() => {
    mockUpdateCustomFields.mockReset().mockResolvedValue({});
    vi.mocked(useDevice).mockReturnValue({
      device: null,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
  });

  describe("loading state", () => {
    it("renders a spinner while loading", () => {
      vi.mocked(useDevice).mockReturnValue({
        device: null,
        isLoading: true,
        error: null,
        refetch: vi.fn(),
      });
      renderPage();
      expect(document.querySelector(".animate-spin")).toBeInTheDocument();
    });
  });

  describe("device data", () => {
    beforeEach(() => {
      vi.mocked(useDevice).mockReturnValue({
        device: makeDevice(),
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
    });

    it("renders the device name as a heading", () => {
      renderPage();
      expect(screen.getByRole("heading", { name: "my-device" })).toBeInTheDocument();
    });

    it("renders the MAC address", () => {
      renderPage();
      expect(screen.getByText("aa:bb:cc:dd:ee:ff")).toBeInTheDocument();
    });

    it("renders the operating system", () => {
      renderPage();
      expect(screen.getByText("Ubuntu 22.04 LTS")).toBeInTheDocument();
    });

    it('renders the "Custom Fields" section label', () => {
      renderPage();
      expect(screen.getByText("Custom Fields")).toBeInTheDocument();
    });
  });

  describe("custom fields section", () => {
    it("renders key-value pairs when custom fields are present", () => {
      vi.mocked(useDevice).mockReturnValue({
        device: makeDevice({ custom_fields: { env: "production", owner: "team-a" } }),
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderPage();
      expect(screen.getByText("env:")).toBeInTheDocument();
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("owner:")).toBeInTheDocument();
      expect(screen.getByText("team-a")).toBeInTheDocument();
    });

    it("renders the add form inputs", () => {
      vi.mocked(useDevice).mockReturnValue({
        device: makeDevice({ custom_fields: {} }),
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderPage();
      expect(screen.getByPlaceholderText("key")).toBeInTheDocument();
      expect(screen.getByPlaceholderText("value")).toBeInTheDocument();
    });

    it("shows delete confirmation when the remove button is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useDevice).mockReturnValue({
        device: makeDevice({ custom_fields: { env: "production" } }),
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderPage();

      // Find the field row by key text and click the adjacent remove button
      const keyEl = screen.getByText("env:");
      const fieldRow = keyEl.closest("div")!.parentElement!;
      const xBtn = within(fieldRow).getByRole("button");
      await user.click(xBtn);

      expect(screen.getByText("Remove?")).toBeInTheDocument();
      expect(screen.getByText("Yes")).toBeInTheDocument();
      expect(screen.getByText("No")).toBeInTheDocument();
    });

    it("hides the confirmation when 'No' is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useDevice).mockReturnValue({
        device: makeDevice({ custom_fields: { env: "production" } }),
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderPage();

      const keyEl = screen.getByText("env:");
      const fieldRow = keyEl.closest("div")!.parentElement!;
      const xBtn = within(fieldRow).getByRole("button");
      await user.click(xBtn);
      await user.click(screen.getByText("No"));

      expect(screen.queryByText("Remove?")).not.toBeInTheDocument();
    });

    it("calls mutation without the deleted key when 'Yes' is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useDevice).mockReturnValue({
        device: makeDevice({ custom_fields: { env: "production", owner: "team-a" } }),
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderPage();

      const keyEl = screen.getByText("env:");
      const fieldRow = keyEl.closest("div")!.parentElement!;
      const xBtn = within(fieldRow).getByRole("button");
      await user.click(xBtn);
      await user.click(screen.getByText("Yes"));

      expect(mockUpdateCustomFields).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            custom_fields: { owner: "team-a" },
          }),
        }),
      );
    });

    it("calls mutation with new field when add form is submitted via Enter key", async () => {
      const user = userEvent.setup();
      vi.mocked(useDevice).mockReturnValue({
        device: makeDevice({ custom_fields: {} }),
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderPage();

      await user.type(screen.getByPlaceholderText("key"), "region");
      await user.type(screen.getByPlaceholderText("value"), "us-east{Enter}");

      expect(mockUpdateCustomFields).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            custom_fields: { region: "us-east" },
          }),
        }),
      );
    });

    it("shows an error when trying to add a duplicate key", async () => {
      const user = userEvent.setup();
      vi.mocked(useDevice).mockReturnValue({
        device: makeDevice({ custom_fields: { env: "production" } }),
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderPage();

      await user.type(screen.getByPlaceholderText("key"), "env");
      await user.type(screen.getByPlaceholderText("value"), "staging{Enter}");

      expect(screen.getByText("This key already exists.")).toBeInTheDocument();
      expect(mockUpdateCustomFields).not.toHaveBeenCalled();
    });
  });
});
