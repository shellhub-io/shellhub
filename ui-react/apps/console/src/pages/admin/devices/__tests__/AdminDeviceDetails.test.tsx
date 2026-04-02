import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { NormalizedDevice } from "../../../../hooks/useAdminDevices";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("../../../../hooks/useAdminDevices", () => ({
  useAdminDevice: vi.fn(),
}));

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useParams: () => ({ uid: "test-uid" }) };
});

// CopyButton relies on ClipboardProvider context and calls showModal() via
// BaseDialog, which is not supported in jsdom. Mock it to a simple no-op.
vi.mock("@/components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button type="button" aria-label={`Copy ${text}`} />
  ),
}));

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { useAdminDevice } from "../../../../hooks/useAdminDevices";
import AdminDeviceDetails from "../AdminDeviceDetails";

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeDevice(
  overrides: Partial<NormalizedDevice> = {},
): NormalizedDevice {
  return {
    uid: "test-uid",
    name: "my-device",
    status: "accepted",
    online: true,
    namespace: "my-namespace",
    tenant_id: "tenant-abc",
    tags: ["production", "web"],
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
    remote_addr: "192.168.1.100",
    public_key: null,
    ...overrides,
  } as NormalizedDevice;
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminDeviceDetails />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AdminDeviceDetails", () => {
  beforeEach(() => {
    vi.mocked(useAdminDevice).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useAdminDevice>);
  });

  describe("loading state", () => {
    it('renders an sr-only "Loading device details" message while loading', () => {
      vi.mocked(useAdminDevice).mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      } as ReturnType<typeof useAdminDevice>);

      renderPage();

      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading device details")).toBeInTheDocument();
    });
  });

  describe("not-found / error state", () => {
    it('renders "Device not found" when no data and no loading', () => {
      renderPage();
      expect(screen.getByText("Device not found")).toBeInTheDocument();
    });

    it('renders "Device not found" when the hook returns an error', () => {
      vi.mocked(useAdminDevice).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error("404 Not Found"),
      } as ReturnType<typeof useAdminDevice>);

      renderPage();
      expect(screen.getByText("Device not found")).toBeInTheDocument();
    });

    it('renders a "Back to devices" link in the not-found state', () => {
      renderPage();
      expect(
        screen.getByRole("link", { name: "Back to devices" }),
      ).toBeInTheDocument();
    });
  });

  describe("device data", () => {
    beforeEach(() => {
      vi.mocked(useAdminDevice).mockReturnValue({
        data: makeDevice(),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminDevice>);
    });

    it("renders the device name as the main heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "my-device" }),
      ).toBeInTheDocument();
    });

    it("renders the device UID", () => {
      renderPage();
      expect(screen.getByText("test-uid")).toBeInTheDocument();
    });

    it("renders the MAC address", () => {
      renderPage();
      expect(screen.getByText("aa:bb:cc:dd:ee:ff")).toBeInTheDocument();
    });

    it("renders the operating system name", () => {
      renderPage();
      expect(screen.getByText("Ubuntu 22.04 LTS")).toBeInTheDocument();
    });

    it("renders the tenant ID", () => {
      renderPage();
      expect(screen.getByText("tenant-abc")).toBeInTheDocument();
    });

    it("renders the status chip", () => {
      renderPage();
      // DeviceStatusChip renders the label text
      expect(screen.getByText("Accepted")).toBeInTheDocument();
    });

    it("renders device tags", () => {
      renderPage();
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("web")).toBeInTheDocument();
    });

    it('renders "No tags" when device has no tags', () => {
      vi.mocked(useAdminDevice).mockReturnValue({
        data: makeDevice({ tags: [] }),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminDevice>);

      renderPage();
      expect(screen.getByText("No tags")).toBeInTheDocument();
    });

    it("renders the public key section when present", () => {
      vi.mocked(useAdminDevice).mockReturnValue({
        data: makeDevice({ public_key: "ssh-rsa AAAAB3NzaC1yc2E..." }),
        isLoading: false,
        error: null,
      } as ReturnType<typeof useAdminDevice>);

      renderPage();
      expect(
        screen.getByText("ssh-rsa AAAAB3NzaC1yc2E..."),
      ).toBeInTheDocument();
    });

    it("does not render the public key section when absent", () => {
      renderPage();
      expect(screen.queryByText(/ssh-rsa/)).not.toBeInTheDocument();
    });

    it("renders the namespace link", () => {
      renderPage();
      expect(
        screen.getByRole("link", { name: "my-namespace" }),
      ).toBeInTheDocument();
    });

    it("renders a breadcrumb back-link to devices", () => {
      renderPage();
      expect(
        screen.getByRole("navigation", { name: "Breadcrumb" }),
      ).toBeInTheDocument();
    });
  });
});
