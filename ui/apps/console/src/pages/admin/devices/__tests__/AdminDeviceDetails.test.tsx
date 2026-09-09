import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import AdminDeviceDetails from "../AdminDeviceDetails";
import type { Device } from "@/client/model";

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

function setDevice(overrides: Partial<Device> = {}) {
  server.use(
    http.get("*/admin/api/devices/:uid", () =>
      HttpResponse.json(makeDevice(overrides)),
    ),
  );
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
  setDevice();
});

describe("AdminDeviceDetails", () => {
  describe("loading state", () => {
    it('announces "Loading device details" while loading', () => {
      server.use(
        http.get("*/admin/api/devices/:uid", () => new Promise(() => {})),
      );
      renderPage();
      expect(
        screen.getByRole("status", { name: "Loading device details" }),
      ).toBeInTheDocument();
    });
  });

  describe("not-found / error state", () => {
    it('renders "Device not found" when no data and no loading', async () => {
      server.use(
        http.get("*/admin/api/devices/:uid", () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      );
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("Device not found")).toBeInTheDocument();
      });
    });
  });

  describe("device data", () => {
    it("renders the device's fields", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("heading", { name: "my-device" }),
        ).toBeInTheDocument();
      });
      expect(screen.getByText("test-uid")).toBeInTheDocument();
      expect(screen.getByText("aa:bb:cc:dd:ee:ff")).toBeInTheDocument();
      expect(screen.getByText("Ubuntu 22.04 LTS")).toBeInTheDocument();
      expect(screen.getByText("tenant-abc")).toBeInTheDocument();
      expect(screen.getByText("Accepted")).toBeInTheDocument();
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("web")).toBeInTheDocument();
      expect(
        screen.getByRole("link", { name: "my-namespace" }),
      ).toBeInTheDocument();
    });

    it('renders "No tags" when device has no tags', async () => {
      setDevice({ tags: [] });
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("No tags")).toBeInTheDocument();
      });
    });

    it("renders the public key section when present", async () => {
      setDevice({ public_key: "ssh-rsa AAAAB3NzaC1yc2E..." });
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByText("ssh-rsa AAAAB3NzaC1yc2E..."),
        ).toBeInTheDocument();
      });
    });
  });
});
