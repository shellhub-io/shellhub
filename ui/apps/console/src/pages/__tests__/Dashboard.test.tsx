import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { mockDevice, mockNamespace } from "@/tests/factories";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { seedAuthStore } from "@/tests/seedAuthStore";
import Dashboard from "../Dashboard";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getStatusDevices: vi.fn(),
    getNamespace: vi.fn(),
    getSessions: vi.fn(),
    getContainers: vi.fn(),
  }),
);

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

function renderDashboard() {
  return render(<Dashboard />, {
    wrapper: createTestWrapper({ initialEntries: ["/dashboard"] }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(mockNamespace()));
  sdk.getSessions.mockResolvedValue(paginatedResponse([]));
  sdk.getContainers.mockResolvedValue(paginatedResponse([]));
  sdk.getStatusDevices.mockResolvedValue(
    mockSdkResponse({
      registered_devices: 4,
      online_devices: 2,
      pending_devices: 3,
      rejected_devices: 0,
    }),
  );
});

describe("Dashboard", () => {
  it("keeps the first-run wizard away from a namespace whose only device is a pending container", async () => {
    sdk.getStatusDevices.mockResolvedValue(
      mockSdkResponse({
        registered_devices: 0,
        online_devices: 0,
        pending_devices: 0,
        rejected_devices: 0,
      }),
    );
    sdk.getContainers.mockResolvedValue(paginatedResponse([mockDevice()], 1));

    renderDashboard();

    expect(await screen.findByText(/pending devices/i)).toBeInTheDocument();
    expect(
      screen.queryByText(/let's connect your first device/i),
    ).not.toBeInTheDocument();
  });

  it("shows the first-run wizard for a namespace with nothing in it", async () => {
    sdk.getStatusDevices.mockResolvedValue(
      mockSdkResponse({
        registered_devices: 0,
        online_devices: 0,
        pending_devices: 0,
        rejected_devices: 0,
      }),
    );

    renderDashboard();

    expect(await screen.findByText(/let's connect your first device/i)).toBeInTheDocument();
  });

  it("sends the pending devices card to the pending review queue", async () => {
    renderDashboard();

    const link = await screen.findByRole("link", { name: /review pending/i });

    expect(link).toHaveAttribute("href", "/pending-devices");
  });
});
