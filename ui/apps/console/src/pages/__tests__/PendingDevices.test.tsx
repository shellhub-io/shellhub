import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { mockDevice, mockInstallKey, mockNamespace } from "@/tests/factories";
import { makeSdkError, mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { seedAuthStore } from "@/tests/seedAuthStore";
import PendingDevices from "../PendingDevices";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getDevices: vi.fn(),
    installKeyList: vi.fn(),
    getNamespace: vi.fn(),
    acceptDevice: vi.fn(),
    updateDeviceStatus: vi.fn(),
  }),
);

function pendingDevice(overrides: Parameters<typeof mockDevice>[0] = {}) {
  return mockDevice({
    uid: "pending-uid-1",
    name: "waiting-host",
    status: "pending",
    online: false,
    install_key_id: "key-digest-1",
    ...overrides,
  });
}

function renderPage() {
  return render(<PendingDevices />, {
    wrapper: createTestWrapper({ initialEntries: ["/pending-devices"] }),
  });
}

async function deviceRow(name = "waiting-host") {
  return within(await screen.findByRole("row", { name: new RegExp(name) }));
}

async function openConfirmation(name: RegExp) {
  const user = userEvent.setup();
  await user.click((await deviceRow()).getByRole("button", { name }));

  return { user, dialog: await screen.findByRole("dialog") };
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getDevices.mockResolvedValue(paginatedResponse([pendingDevice()]));
  sdk.installKeyList.mockResolvedValue(
    paginatedResponse([
      mockInstallKey({ id: "key-digest-1", name: "fleet-key" }),
    ]),
  );
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(mockNamespace()));
  sdk.acceptDevice.mockResolvedValue(mockSdkResponse(undefined));
  sdk.updateDeviceStatus.mockResolvedValue(mockSdkResponse(undefined));
});

describe("Pending devices", () => {
  it("asks the API only for devices awaiting review", async () => {
    renderPage();

    await screen.findByText("waiting-host");

    expect(sdk.getDevices).toHaveBeenCalledWith(
      expect.objectContaining({
        query: expect.objectContaining({ status: "pending" }),
      }),
    );
  });

  it("names the install key each device registered with", async () => {
    renderPage();

    expect(
      await screen.findByRole("link", { name: "fleet-key" }),
    ).toHaveAttribute("href", "/install-keys/key-digest-1/activity");
  });

  it("falls back to a dash when the install key is unknown", async () => {
    sdk.installKeyList.mockResolvedValue(paginatedResponse([]));

    renderPage();

    expect((await deviceRow()).getByText("—")).toBeInTheDocument();
  });

  it("accepts a device only after the confirmation is confirmed", async () => {
    renderPage();

    const { user, dialog } = await openConfirmation(/^accept$/i);

    expect(within(dialog).getByText(/waiting-host/)).toBeInTheDocument();
    expect(sdk.acceptDevice).not.toHaveBeenCalled();

    await user.click(within(dialog).getByRole("button", { name: /^accept$/i }));

    expect(sdk.acceptDevice).toHaveBeenCalledWith(
      expect.objectContaining({ path: { uid: "pending-uid-1" } }),
    );
  });

  it("rejects a device only after the confirmation is confirmed", async () => {
    renderPage();

    const { user, dialog } = await openConfirmation(/^reject$/i);

    expect(within(dialog).getByText(/waiting-host/)).toBeInTheDocument();
    expect(sdk.updateDeviceStatus).not.toHaveBeenCalled();

    await user.click(within(dialog).getByRole("button", { name: /^reject$/i }));

    expect(sdk.updateDeviceStatus).toHaveBeenCalledWith(
      expect.objectContaining({
        path: { uid: "pending-uid-1", status: "reject" },
      }),
    );
  });

  it("marks the decision unavailable for a member without permission", async () => {
    seedAuthStore({ role: "observer" });

    renderPage();

    const restricted = (await deviceRow()).getAllByTitle(
      /don't have permission/i,
    );

    expect(restricted.length).toBeGreaterThan(0);
    for (const wrapper of restricted) {
      expect(wrapper).toHaveAttribute("aria-disabled", "true");
    }
  });

  it("reports a failure to load rather than an empty queue", async () => {
    sdk.getDevices.mockRejectedValue(makeSdkError(500));

    renderPage();

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(
      screen.queryByText(/no devices are waiting for review/i),
    ).not.toBeInTheDocument();
  });

  it("says the queue is clear when nothing is pending", async () => {
    sdk.getDevices.mockResolvedValue(paginatedResponse([]));

    renderPage();

    expect(
      await screen.findByText(/no devices are waiting for review/i),
    ).toBeInTheDocument();
  });
});
