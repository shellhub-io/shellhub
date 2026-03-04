import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import { useAuthStore } from "@/stores/authStore";
import { useStatsStore } from "@/stores/statsStore";

// Mock the API (used internally by statsStore)
vi.mock("@/api/stats", () => ({
  getStats: vi.fn(),
}));

// Mock welcomeState utilities so we can observe calls
vi.mock("@/utils/welcomeState", () => ({
  hasSeenWelcome: vi.fn(),
  markWelcomeSeen: vi.fn(),
}));

// Mock WelcomeWizard so we don't render the full modal in these unit tests
vi.mock("../WelcomeWizard", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open ? <div data-testid="welcome-wizard" onClick={onClose} /> : null,
}));

import { getStats } from "@/api/stats";
import { hasSeenWelcome, markWelcomeSeen } from "@/utils/welcomeState";
import WelcomeWizardTrigger from "../WelcomeWizardTrigger";

const mockGetStats = vi.mocked(getStats);
const mockHasSeenWelcome = vi.mocked(hasSeenWelcome);
const mockMarkWelcomeSeen = vi.mocked(markWelcomeSeen);

const zeroStats = {
  registered_devices: 0,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ tenant: "tenant-abc" } as never);
  useStatsStore.setState({ stats: null });
  mockHasSeenWelcome.mockReturnValue(false);
});

afterEach(cleanup);

describe("WelcomeWizardTrigger", () => {
  describe("when tenant has already seen welcome", () => {
    it("renders nothing and does not call getStats", () => {
      mockHasSeenWelcome.mockReturnValue(true);

      render(<WelcomeWizardTrigger />);

      expect(mockGetStats).not.toHaveBeenCalled();
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when tenant has not seen welcome and has zero devices", () => {
    it("shows the wizard", async () => {
      mockGetStats.mockResolvedValue(zeroStats);

      render(<WelcomeWizardTrigger />);

      await waitFor(() => {
        expect(screen.getByTestId("welcome-wizard")).toBeInTheDocument();
      });
    });

    it("calls markWelcomeSeen with the tenant id when wizard is closed", async () => {
      mockGetStats.mockResolvedValue(zeroStats);

      render(<WelcomeWizardTrigger />);

      // Wait for wizard to appear
      const wizard = await screen.findByTestId("welcome-wizard");

      // markWelcomeSeen should NOT be called yet (only on close)
      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();

      // Simulate closing the wizard via the onClose prop
      wizard.click();

      await waitFor(() => {
        expect(mockMarkWelcomeSeen).toHaveBeenCalledWith("tenant-abc");
      });
    });

    it("hides the wizard after it is closed", async () => {
      mockGetStats.mockResolvedValue(zeroStats);

      render(<WelcomeWizardTrigger />);

      const wizard = await screen.findByTestId("welcome-wizard");
      wizard.click();

      await waitFor(() => {
        expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
      });
    });
  });

  describe("when tenant has devices", () => {
    it("does not show the wizard when there are registered devices", async () => {
      mockGetStats.mockResolvedValue({ ...zeroStats, registered_devices: 1 });

      render(<WelcomeWizardTrigger />);

      await waitFor(() => {
        expect(mockGetStats).toHaveBeenCalled();
      });

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });

    it("does not show the wizard when there are pending devices", async () => {
      mockGetStats.mockResolvedValue({ ...zeroStats, pending_devices: 2 });

      render(<WelcomeWizardTrigger />);

      await waitFor(() => {
        expect(mockGetStats).toHaveBeenCalled();
      });

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });

    it("does not show the wizard when there are rejected devices", async () => {
      mockGetStats.mockResolvedValue({ ...zeroStats, rejected_devices: 1 });

      render(<WelcomeWizardTrigger />);

      await waitFor(() => {
        expect(mockGetStats).toHaveBeenCalled();
      });

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });

    it("does not call markWelcomeSeen when there are devices", async () => {
      mockGetStats.mockResolvedValue({ ...zeroStats, registered_devices: 5 });

      render(<WelcomeWizardTrigger />);

      await waitFor(() => {
        expect(mockGetStats).toHaveBeenCalled();
      });

      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();
    });
  });

  describe("when getStats throws", () => {
    it("does not show the wizard", async () => {
      mockGetStats.mockRejectedValue(new Error("network error"));

      render(<WelcomeWizardTrigger />);

      await waitFor(() => {
        expect(mockGetStats).toHaveBeenCalled();
      });

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when tenant is null", () => {
    it("does not call getStats", () => {
      useAuthStore.setState({ tenant: null } as never);

      render(<WelcomeWizardTrigger />);

      expect(mockGetStats).not.toHaveBeenCalled();
    });
  });
});
