import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import AdminDeviceDetails from "../AdminDeviceDetails";
import type { Device } from "@/client";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getDeviceAdmin: vi.fn(),
  }),
);

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useParams: () => ({ uid: "test-uid" }) };
});

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

function makeDevice(overrides: Partial<Device> = {}): Device {
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
  } as Device;
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminDeviceDetails />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  sdk.getDeviceAdmin.mockResolvedValue(mockSdkResponse(makeDevice()));
});

describe("AdminDeviceDetails", () => {
  describe("loading state", () => {
    it('announces "Loading device details" while loading', () => {
      sdk.getDeviceAdmin.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(
        screen.getByRole("status", { name: "Loading device details" }),
      ).toBeInTheDocument();
    });
  });

  describe("not-found / error state", () => {
    it('renders "Device not found" when no data and no loading', async () => {
      sdk.getDeviceAdmin.mockRejectedValue(makeSdkError(404));
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Device not found")).toBeInTheDocument();
      });
    });

    it('renders "Device not found" when the query returns an error', async () => {
      sdk.getDeviceAdmin.mockRejectedValue(makeSdkError(500));
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Device not found")).toBeInTheDocument();
      });
    });

    it('renders a "Back to devices" link in the not-found state', async () => {
      sdk.getDeviceAdmin.mockRejectedValue(makeSdkError(404));
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("link", { name: "Back to devices" }),
        ).toBeInTheDocument();
      });
    });
  });

  describe("device data", () => {
    it("renders the device name as the main heading", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: "my-device" }),
        ).toBeInTheDocument();
      });
    });

    it("renders the device UID", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("test-uid")).toBeInTheDocument();
      });
    });

    it("renders the MAC address", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("aa:bb:cc:dd:ee:ff")).toBeInTheDocument();
      });
    });

    it("renders the operating system name", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Ubuntu 22.04 LTS")).toBeInTheDocument();
      });
    });

    it("renders the tenant ID", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("tenant-abc")).toBeInTheDocument();
      });
    });

    it("renders the status chip", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Accepted")).toBeInTheDocument();
      });
    });

    it("renders device tags", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("production")).toBeInTheDocument();
      });
      expect(screen.getByText("web")).toBeInTheDocument();
    });

    it('renders "No tags" when device has no tags', async () => {
      sdk.getDeviceAdmin.mockResolvedValue(
        mockSdkResponse(makeDevice({ tags: [] })),
      );
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("No tags")).toBeInTheDocument();
      });
    });

    it("renders the public key section when present", async () => {
      sdk.getDeviceAdmin.mockResolvedValue(
        mockSdkResponse(
          makeDevice({ public_key: "ssh-rsa AAAAB3NzaC1yc2E..." }),
        ),
      );
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByText("ssh-rsa AAAAB3NzaC1yc2E..."),
        ).toBeInTheDocument();
      });
    });

    it("does not render the public key section when absent", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: "my-device" }),
        ).toBeInTheDocument();
      });
      expect(screen.queryByText(/ssh-rsa/)).not.toBeInTheDocument();
    });

    it("renders the namespace link", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("link", { name: "my-namespace" }),
        ).toBeInTheDocument();
      });
    });
  });
});
