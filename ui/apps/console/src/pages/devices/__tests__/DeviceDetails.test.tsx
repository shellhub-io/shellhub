import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import type { Device } from "@/client";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import {
  mockDevice as mockDeviceFactory,
  mockNamespace,
} from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getDevice: vi.fn(),
    updateDevice: vi.fn(),
    createTag: vi.fn(),
    pushTagToDevice: vi.fn(),
    pullTagFromDevice: vi.fn(),
    deleteDevice: vi.fn(),
    setDeviceCustomField: vi.fn(),
    deleteDeviceCustomField: vi.fn(),
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
    getTags: vi.fn(),
    installKeyList: vi.fn(),
  }),
);

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (
    sel: (s: { sessions: []; restore: () => void }) => unknown,
  ) => sel({ sessions: [], restore: vi.fn() }),
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
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

vi.mock("@/components/common/ActionDialog", () => ({
  default: () => null,
}));

const mockRequestAction = vi.fn();
let capturedOnSuccess: ((action: string) => void) | undefined;

vi.mock("@/hooks/useActionDialog", () => ({
  useActionDialog: (opts?: { onSuccess?: (action: string) => void }) => {
    capturedOnSuccess = opts?.onSuccess;
    return {
      action: undefined,
      actionKey: "closed",
      requestAction: mockRequestAction,
      close: vi.fn(),
      handleSuccess: vi.fn(),
    };
  },
}));

vi.mock("@/utils/date", () => ({
  formatRelative: () => "just now",
  formatDateFull: () => "Jan 15, 2024",
}));

vi.mock("@/utils/sshid", () => ({
  buildSshid: (ns: string, name: string) => `${ns}.${name}@localhost`,
}));

const mockNavigate = vi.fn();

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return {
    ...actual,
    useParams: () => ({ uid: "test-uid" }),
    useNavigate: () => mockNavigate,
    useSearchParams: () => [new URLSearchParams(), vi.fn()],
  };
});

import DeviceDetails from "@/pages/DeviceDetails";

function makeDevice(overrides: Partial<Device> = {}): Device {
  return mockDeviceFactory({
    uid: "test-uid",
    name: "my-device",
    status: "accepted",
    online: true,
    last_seen: "2024-01-15T10:00:00.000Z",
    created_at: "2023-06-01T08:00:00.000Z",
    info: {
      id: "ubuntu",
      pretty_name: "Ubuntu 22.04 LTS",
      arch: "x86_64",
      platform: "native",
      version: "0.14.0",
    },
    ...overrides,
  });
}

function renderPage() {
  return render(<DeviceDetails />, {
    wrapper: createTestWrapper({ initialEntries: ["/devices/test-uid"] }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getDevice.mockResolvedValue(mockSdkResponse(null));
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(mockNamespace()));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
  sdk.getTags.mockResolvedValue(mockSdkResponse([]));
  sdk.installKeyList.mockResolvedValue(paginatedResponse([]));
  sdk.updateDevice.mockResolvedValue(mockSdkResponse(undefined));
  sdk.createTag.mockResolvedValue(mockSdkResponse(undefined));
  sdk.pushTagToDevice.mockResolvedValue(mockSdkResponse(undefined));
  sdk.pullTagFromDevice.mockResolvedValue(mockSdkResponse(undefined));
  sdk.deleteDevice.mockResolvedValue(mockSdkResponse(undefined));
  sdk.setDeviceCustomField.mockResolvedValue(mockSdkResponse(undefined));
  sdk.deleteDeviceCustomField.mockResolvedValue(mockSdkResponse(undefined));
  mockRequestAction.mockReset();
  mockNavigate.mockReset();
  capturedOnSuccess = undefined;
});

describe("DeviceDetails", () => {
  describe("loading state", () => {
    it("renders a spinner while loading", () => {
      sdk.getDevice.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(document.querySelector(".animate-spin")).toBeInTheDocument();
    });
  });

  describe("device data", () => {
    beforeEach(() => {
      sdk.getDevice.mockResolvedValue(mockSdkResponse(makeDevice()));
    });

    it("renders the device name as a heading", async () => {
      renderPage();
      expect(
        await screen.findByRole("heading", { name: "my-device" }),
      ).toBeInTheDocument();
    });

    it("renders the MAC address", async () => {
      renderPage();
      expect(await screen.findByText("aa:bb:cc:dd:ee:ff")).toBeInTheDocument();
    });

    it("renders the operating system", async () => {
      renderPage();
      expect(await screen.findByText("Ubuntu 22.04 LTS")).toBeInTheDocument();
    });

    it('renders the "Custom Fields" section label', async () => {
      renderPage();
      expect(await screen.findByText("Custom Fields")).toBeInTheDocument();
    });
  });

  describe("custom fields section", () => {
    it("renders key-value pairs when custom fields are present", async () => {
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(
          makeDevice({ custom_fields: { env: "production", owner: "team-a" } }),
        ),
      );
      renderPage();
      expect(await screen.findByText("env:")).toBeInTheDocument();
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("owner:")).toBeInTheDocument();
      expect(screen.getByText("team-a")).toBeInTheDocument();
    });

    it("renders the add form inputs", async () => {
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ custom_fields: {} })),
      );
      renderPage();
      expect(await screen.findByPlaceholderText("key")).toBeInTheDocument();
      expect(screen.getByPlaceholderText("value")).toBeInTheDocument();
    });

    it("shows delete confirmation when the remove button is clicked", async () => {
      const user = userEvent.setup();
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ custom_fields: { env: "production" } })),
      );
      renderPage();
      await screen.findByText("env:");

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
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ custom_fields: { env: "production" } })),
      );
      renderPage();
      await screen.findByText("env:");

      const keyEl = screen.getByText("env:");
      const fieldRow = keyEl.closest("div")!.parentElement!;
      const xBtn = within(fieldRow).getByRole("button");
      await user.click(xBtn);
      await user.click(screen.getByText("No"));

      expect(screen.queryByText("Remove?")).not.toBeInTheDocument();
    });

    it("calls deleteDeviceCustomField when 'Yes' is clicked", async () => {
      const user = userEvent.setup();
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(
          makeDevice({
            custom_fields: { env: "production", owner: "team-a" },
          }),
        ),
      );
      renderPage();
      await screen.findByText("env:");

      const keyEl = screen.getByText("env:");
      const fieldRow = keyEl.closest("div")!.parentElement!;
      const xBtn = within(fieldRow).getByRole("button");
      await user.click(xBtn);
      await user.click(screen.getByText("Yes"));

      expect(sdk.deleteDeviceCustomField).toHaveBeenCalledWith(
        expect.objectContaining({
          path: expect.objectContaining({ uid: "test-uid", key: "env" }),
        }),
      );
    });

    it("calls setDeviceCustomField when add form is submitted via Enter key", async () => {
      const user = userEvent.setup();
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ custom_fields: {} })),
      );
      renderPage();
      await screen.findByPlaceholderText("key");

      await user.type(screen.getByPlaceholderText("key"), "region");
      await user.type(screen.getByPlaceholderText("value"), "us-east{Enter}");

      expect(sdk.setDeviceCustomField).toHaveBeenCalledWith(
        expect.objectContaining({
          path: expect.objectContaining({ uid: "test-uid", key: "region" }),
          body: { value: "us-east" },
        }),
      );
    });

    it("shows an error when trying to add a duplicate key", async () => {
      const user = userEvent.setup();
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ custom_fields: { env: "production" } })),
      );
      renderPage();
      await screen.findByPlaceholderText("key");

      await user.type(screen.getByPlaceholderText("key"), "env");
      await user.type(screen.getByPlaceholderText("value"), "staging{Enter}");

      expect(screen.getByText("This key already exists.")).toBeInTheDocument();
      expect(sdk.setDeviceCustomField).not.toHaveBeenCalled();
    });
  });

  describe("action buttons delegate to useDeviceActions", () => {
    it("calls requestAction('accept') when Accept is clicked on a pending device", async () => {
      const user = userEvent.setup();
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ status: "pending", online: false })),
      );
      renderPage();

      await user.click(await screen.findByRole("button", { name: /Accept/i }));

      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({ uid: "test-uid" }),
        "accept",
      );
    });

    it("calls requestAction('reject') when Reject is clicked on a pending device", async () => {
      const user = userEvent.setup();
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ status: "pending", online: false })),
      );
      renderPage();

      await user.click(await screen.findByRole("button", { name: /Reject/i }));

      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({ uid: "test-uid" }),
        "reject",
      );
    });

    it("calls requestAction('remove') when Remove is clicked on a rejected device", async () => {
      const user = userEvent.setup();
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ status: "rejected", online: false })),
      );
      renderPage();

      await user.click(await screen.findByRole("button", { name: /Remove/i }));

      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({ uid: "test-uid" }),
        "remove",
      );
    });

    it("calls requestAction('remove') when the Delete device trash button is clicked on an accepted device", async () => {
      const user = userEvent.setup();
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ status: "accepted", online: true })),
      );
      renderPage();

      await user.click(
        await screen.findByRole("button", { name: "Delete device" }),
      );

      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({ uid: "test-uid" }),
        "remove",
      );
    });
  });

  describe("onSuccess callback wiring", () => {
    it("navigates to /devices when onSuccess is called with action 'remove'", async () => {
      sdk.getDevice.mockResolvedValue(mockSdkResponse(makeDevice()));
      renderPage();
      await screen.findByRole("heading", { name: "my-device" });

      expect(capturedOnSuccess).toBeDefined();
      capturedOnSuccess!("remove");

      expect(mockNavigate).toHaveBeenCalledWith("/devices");
    });

    it("does NOT navigate when onSuccess is called with a non-remove action", async () => {
      sdk.getDevice.mockResolvedValue(
        mockSdkResponse(makeDevice({ status: "pending", online: false })),
      );
      renderPage();
      await screen.findByRole("heading", { name: "my-device" });

      expect(capturedOnSuccess).toBeDefined();
      capturedOnSuccess!("accept");

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });
});
