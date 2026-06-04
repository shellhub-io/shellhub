import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import { useAuthStore } from "@/stores/authStore";

vi.mock("@/hooks/useStats", () => ({
  useStats: vi.fn(),
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

import { useStats } from "@/hooks/useStats";
import { hasSeenWelcome, markWelcomeSeen } from "@/utils/welcomeState";
import WelcomeWizardTrigger from "../WelcomeWizardTrigger";

const mockUseStats = vi.mocked(useStats);
const mockHasSeenWelcome = vi.mocked(hasSeenWelcome);
const mockMarkWelcomeSeen = vi.mocked(markWelcomeSeen);

const zeroStats = {
  registered_devices: 0,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const mockRefetch = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ tenant: "tenant-abc" } as never);
  mockUseStats.mockReturnValue({
    stats: null,
    isLoading: true,
    error: null,
    refetch: mockRefetch,
  });
  mockHasSeenWelcome.mockReturnValue(false);
});

afterEach(cleanup);

describe("WelcomeWizardTrigger", () => {
  describe("when tenant has already seen welcome", () => {
    it("renders nothing", () => {
      mockHasSeenWelcome.mockReturnValue(true);
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when tenant has not seen welcome and has zero devices", () => {
    it("shows the wizard", () => {
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(screen.getByTestId("welcome-wizard")).toBeInTheDocument();
    });

    it("calls markWelcomeSeen with the tenant id when wizard is closed", async () => {
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      const wizard = screen.getByTestId("welcome-wizard");

      // markWelcomeSeen should NOT be called yet (only on close)
      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();

      // Simulate closing the wizard via the onClose prop
      wizard.click();

      await waitFor(() => {
        expect(mockMarkWelcomeSeen).toHaveBeenCalledWith("tenant-abc");
      });
    });

    it("hides the wizard after it is closed", async () => {
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      const wizard = screen.getByTestId("welcome-wizard");
      wizard.click();

      await waitFor(() => {
        expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
      });
    });

    it("calls refetch when wizard is closed", () => {
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      screen.getByTestId("welcome-wizard").click();

      expect(mockRefetch).toHaveBeenCalled();
    });
  });

  describe("when tenant has devices", () => {
    it("does not show the wizard when there are registered devices", () => {
      mockUseStats.mockReturnValue({
        stats: { ...zeroStats, registered_devices: 1 },
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });

    it("does not show the wizard when there are pending devices", () => {
      mockUseStats.mockReturnValue({
        stats: { ...zeroStats, pending_devices: 2 },
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });

    it("does not show the wizard when there are rejected devices", () => {
      mockUseStats.mockReturnValue({
        stats: { ...zeroStats, rejected_devices: 1 },
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });

    it("does not call markWelcomeSeen when there are devices", () => {
      mockUseStats.mockReturnValue({
        stats: { ...zeroStats, registered_devices: 5 },
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();
    });
  });

  describe("when stats are loading", () => {
    it("does not show the wizard", () => {
      render(<WelcomeWizardTrigger />);

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when tenant is null", () => {
    it("does not show the wizard", () => {
      useAuthStore.setState({ tenant: null } as never);
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });
});
