import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// ── Module mocks ──────────────────────────────────────────────────────────────

// Mock @/env before importing getAcceptDeviceErrorMessage (which reads getConfig)
vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: vi.fn() };
});

const mockMutateAsync = vi.fn();

vi.mock("@/hooks/useDeviceMutations", () => ({
  useAcceptDevice: () => ({ mutateAsync: mockMutateAsync }),
  useRejectDevice: () => ({
    mutateAsync: vi.fn().mockRejectedValue({ status: 500 }),
  }),
  useRemoveDevice: () => ({
    mutateAsync: vi.fn().mockRejectedValue({ status: 500 }),
  }),
}));

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { getConfig, defaultConfig } from "@/env";
import DeviceActionDialog from "../DeviceActionDialog";

// ── Typed mocks ───────────────────────────────────────────────────────────────

const mockGetConfig = vi.mocked(getConfig);

// ── Helpers ───────────────────────────────────────────────────────────────────

const testDevice = { uid: "device-uid-1", name: "my-device" };

function renderDialog({
  action,
  onBillingWarning,
  open = true,
}: {
  action?: "accept" | "reject" | "remove";
  onBillingWarning?: () => void;
  open?: boolean;
} = {}) {
  action ??= "accept";
  const onClose = vi.fn();
  const onSuccess = vi.fn();

  render(
    <DeviceActionDialog
      device={testDevice}
      action={action}
      onClose={onClose}
      onSuccess={onSuccess}
      onBillingWarning={onBillingWarning}
      open={open}
    />,
  );

  return { onClose, onSuccess };
}

async function clickAccept() {
  const user = userEvent.setup();
  const acceptBtn = screen.getByRole("button", { name: "Accept" });
  await user.click(acceptBtn);
}

async function clickConfirm(label: string) {
  const user = userEvent.setup();
  const btn = screen.getByRole("button", { name: label });
  await user.click(btn);
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("DeviceActionDialog — error messages via getAcceptDeviceErrorMessage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetConfig.mockReturnValue({ ...defaultConfig });
  });

  describe("accept action — 402 error", () => {
    it("enterprise (no onBillingWarning): shows license copy, NOT cloud billing copy", async () => {
      // Set enterprise config BEFORE triggering action
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: false,
      });
      mockMutateAsync.mockRejectedValue({ status: 402 });

      renderDialog({ action: "accept" });

      await clickAccept();

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });

      const alertText = screen.getByRole("alert").textContent ?? "";
      expect(alertText).toMatch(/license/i);
      expect(alertText).not.toMatch(/billing|subscription|plan/i);
    });

    it("cloud (onBillingWarning provided): calls onBillingWarning, no error rendered", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: true,
      });
      mockMutateAsync.mockRejectedValue({ status: 402 });

      const onBillingWarning = vi.fn();
      renderDialog({ action: "accept", onBillingWarning });

      await clickAccept();

      await waitFor(() => {
        expect(onBillingWarning).toHaveBeenCalledTimes(1);
      });

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });

    it("community (no enterprise, no cloud): shows generic fallback copy", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: false,
        cloud: false,
      });
      mockMutateAsync.mockRejectedValue({ status: 402 });

      renderDialog({ action: "accept" });

      await clickAccept();

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });

      const alertText = screen.getByRole("alert").textContent ?? "";
      // Must NOT show cloud billing copy
      expect(alertText).not.toMatch(/billing|subscription|plan/i);
      // Must NOT show enterprise license copy
      expect(alertText).not.toMatch(/license/i);
    });
  });

  describe("accept action — other error statuses", () => {
    it("shows a namespace/permission message on 403", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig });
      mockMutateAsync.mockRejectedValue({ status: 403 });

      renderDialog({ action: "accept" });

      await clickAccept();

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });

      const alertText = screen.getByRole("alert").textContent ?? "";
      expect(alertText).toMatch(/namespace|permission/i);
    });

    it("shows a rename message on 409", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig });
      mockMutateAsync.mockRejectedValue({ status: 409 });

      renderDialog({ action: "accept" });

      await clickAccept();

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });

      const alertText = screen.getByRole("alert").textContent ?? "";
      expect(alertText).toMatch(/rename|name|already exists/i);
    });
  });

  describe("accept action — success", () => {
    it("calls onSuccess and onClose when accept succeeds", async () => {
      mockMutateAsync.mockResolvedValue(undefined);

      const { onSuccess, onClose } = renderDialog({ action: "accept" });

      await clickAccept();

      await waitFor(() => expect(onSuccess).toHaveBeenCalledTimes(1));
      expect(onClose).toHaveBeenCalledTimes(1);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("non-accept actions — error does NOT use accept copy", () => {
    it("reject failure shows generic 'Failed to reject device.' copy, NOT accept-specific copy", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig });

      renderDialog({ action: "reject" });

      await clickConfirm("Reject");

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });

      const alertText = screen.getByRole("alert").textContent ?? "";
      expect(alertText).toBe("Failed to reject device.");
    });

    it("remove failure shows generic 'Failed to remove device.' copy, NOT accept-specific copy", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig });

      renderDialog({ action: "remove" });

      await clickConfirm("Remove");

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });

      const alertText = screen.getByRole("alert").textContent ?? "";
      expect(alertText).toBe("Failed to remove device.");
    });
  });
});
