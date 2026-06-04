import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

// ── Feature flag ──────────────────────────────────────────────────────────────

vi.mock("@/env", () => ({ getConfig: vi.fn() }));

// ── Hooks ─────────────────────────────────────────────────────────────────────

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: vi.fn(),
}));

vi.mock("@/hooks/useStats", () => ({
  useStats: vi.fn(),
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: vi.fn(),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (sel: (s: { tenant: string }) => unknown) =>
    sel({ tenant: "tenant-abc" }),
}));

// ── Stub out DeviceChooserDialog so no real dialog logic runs ─────────────────
vi.mock("../DeviceChooserDialog", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open
      ? React.createElement(
          "div",
          { "data-testid": "device-chooser-dialog" },
          React.createElement("button", { onClick: onClose }, "Dismiss"),
        )
      : null,
}));

// ── Post-mock imports ─────────────────────────────────────────────────────────

import { getConfig } from "@/env";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useStats } from "@/hooks/useStats";
import { useNamespace } from "@/hooks/useNamespaces";
import DeviceChooserTrigger from "../DeviceChooserTrigger";

const mockGetConfig = vi.mocked(getConfig);
const mockUseHasPermission = vi.mocked(useHasPermission);
const mockUseStats = vi.mocked(useStats);
const mockUseNamespace = vi.mocked(useNamespace);

// ── Helpers ───────────────────────────────────────────────────────────────────

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function makeStats(registeredDevices: number) {
  return {
    registered_devices: registeredDevices,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 0,
    rejected_devices: 0,
  };
}

function makeNamespace(billingActive: boolean) {
  return {
    billing: { active: billingActive },
  };
}

/** Sets up all hooks for the fully-open scenario and then overrides. */
function setupHooks({
  cloud = true,
  canChoose = true,
  billingActive = false,
  registeredDevices = 4,
  nsLoading = false,
  statsLoading = false,
  namespace,
}: {
  cloud?: boolean;
  canChoose?: boolean;
  billingActive?: boolean;
  registeredDevices?: number;
  nsLoading?: boolean;
  statsLoading?: boolean;
  namespace?: ReturnType<typeof makeNamespace> | null;
} = {}) {
  mockGetConfig.mockReturnValue({ cloud } as ReturnType<typeof getConfig>);
  mockUseHasPermission.mockReturnValue(canChoose);
  mockUseStats.mockReturnValue({
    stats: statsLoading ? null : makeStats(registeredDevices),
    isLoading: statsLoading,
    error: null,
    refetch: vi.fn(),
  });
  const ns = namespace === undefined ? makeNamespace(billingActive) : namespace;
  mockUseNamespace.mockReturnValue({
    namespace: nsLoading ? null : (ns as never),
    isLoading: nsLoading,
    error: null,
    refetch: vi.fn(),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  setupHooks();
});

afterEach(cleanup);

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("DeviceChooserTrigger", () => {
  // ── Gating matrix ────────────────────────────────────────────────────────────

  describe("when cloud=false", () => {
    it("renders nothing without mounting the inner component", () => {
      setupHooks({ cloud: false });
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
      // Inner component never mounts, so useHasPermission is never called
      expect(mockUseHasPermission).not.toHaveBeenCalled();
    });
  });

  describe("when cloud=true but user lacks device:choose permission", () => {
    it("renders nothing", async () => {
      setupHooks({ cloud: true, canChoose: false });
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );
    });
  });

  describe("when cloud=true, owner, billing is active", () => {
    it("renders nothing", async () => {
      setupHooks({ cloud: true, canChoose: true, billingActive: true });
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );
    });
  });

  describe("when cloud=true, owner, billing inactive, registered_devices=3 (boundary)", () => {
    it("renders nothing — limit is strictly greater than 3", async () => {
      setupHooks({
        cloud: true,
        canChoose: true,
        billingActive: false,
        registeredDevices: 3,
      });
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );
    });
  });

  describe("when all conditions are met (cloud, owner, billing inactive, >3 devices)", () => {
    it("auto-opens the dialog on mount", async () => {
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      await waitFor(() =>
        expect(screen.getByTestId("device-chooser-dialog")).toBeInTheDocument(),
      );
    });
  });

  // ── C1 regression: don't render while loading or namespace undefined ─────────

  describe("when namespace is still loading", () => {
    it("renders nothing even if stats indicate over-limit", () => {
      setupHooks({
        nsLoading: true,
        registeredDevices: 4,
        billingActive: false,
      });
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when stats are still loading", () => {
    it("renders nothing", () => {
      setupHooks({ statsLoading: true, billingActive: false });
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when namespace returned is undefined", () => {
    it("renders nothing — billing is unknown until namespace resolves", () => {
      setupHooks({ namespace: null, registeredDevices: 4 });
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when the stats query settled with an error", () => {
    it("renders nothing — overLimit cannot be evaluated without stats", () => {
      setupHooks();
      mockUseStats.mockReturnValue({
        stats: null,
        isLoading: false,
        error: new Error("network failure"),
        refetch: vi.fn(),
      });
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  // ── Dismissal behaviour ──────────────────────────────────────────────────────

  describe("dismissal", () => {
    it("hides the dialog after dismissal", async () => {
      const user = userEvent.setup();
      render(<DeviceChooserTrigger />, { wrapper: createWrapper() });
      await waitFor(() =>
        expect(screen.getByTestId("device-chooser-dialog")).toBeInTheDocument(),
      );
      await user.click(screen.getByRole("button", { name: /dismiss/i }));
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );
    });

    it("does not re-open the dialog after dismissal within the same mount", async () => {
      const user = userEvent.setup();
      const { rerender } = render(<DeviceChooserTrigger />, {
        wrapper: createWrapper(),
      });
      await waitFor(() =>
        expect(screen.getByTestId("device-chooser-dialog")).toBeInTheDocument(),
      );
      await user.click(screen.getByRole("button", { name: /dismiss/i }));
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );

      // Trigger a re-render with the same conditions — dialog must NOT reopen
      rerender(<DeviceChooserTrigger />);
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });

    it("reopens the dialog on a fresh remount after conditions are still true", async () => {
      const user = userEvent.setup();
      const wrapper = createWrapper();

      render(<DeviceChooserTrigger />, { wrapper });
      await waitFor(() =>
        expect(screen.getByTestId("device-chooser-dialog")).toBeInTheDocument(),
      );
      await user.click(screen.getByRole("button", { name: /dismiss/i }));
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );

      // Unmount and remount — fresh component state, so dialog reopens
      cleanup();

      render(<DeviceChooserTrigger />, { wrapper });
      await waitFor(() =>
        expect(screen.getByTestId("device-chooser-dialog")).toBeInTheDocument(),
      );
    });
  });
});
