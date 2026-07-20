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

// Mock WelcomeWizard so we don't render the full modal in these unit tests.
// Exposes close (defer) and dismiss (for good) as separate controls.
vi.mock("../WelcomeWizard", () => ({
  default: ({
    open,
    onClose,
    onDismiss,
  }: {
    open: boolean;
    onClose: () => void;
    onDismiss: () => void;
  }) =>
    open ? (
      <>
        <button
          type="button"
          aria-label="Close wizard"
          data-testid="welcome-wizard"
          onClick={onClose}
        />
        <button
          type="button"
          aria-label="Dismiss wizard"
          data-testid="wizard-dismiss"
          onClick={onDismiss}
        />
      </>
    ) : null,
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

    it("does not call markWelcomeSeen when merely closed", async () => {
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      // Closing defers the wizard for this session; it must not be suppressed
      // for good, so markWelcomeSeen is never called.
      screen.getByTestId("welcome-wizard").click();

      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();
    });

    it("calls markWelcomeSeen with the tenant id when dismissed for good", async () => {
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();

      screen.getByTestId("wizard-dismiss").click();

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

    it("still shows the wizard when a device is only pending (not accepted)", () => {
      mockUseStats.mockReturnValue({
        stats: { ...zeroStats, pending_devices: 2 },
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      // Onboarding isn't done until a device is accepted, so a pending device
      // must not suppress the wizard.
      expect(screen.getByTestId("welcome-wizard")).toBeInTheDocument();
    });

    it("still shows the wizard when a device is only rejected (not accepted)", () => {
      mockUseStats.mockReturnValue({
        stats: { ...zeroStats, rejected_devices: 1 },
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<WelcomeWizardTrigger />);

      expect(screen.getByTestId("welcome-wizard")).toBeInTheDocument();
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

  describe("eligibility is decided once, at page load", () => {
    it("does not reopen when the last device is deleted mid-session", () => {
      // Page loads with an accepted device -> wizard suppressed.
      mockUseStats.mockReturnValue({
        stats: { ...zeroStats, registered_devices: 1 },
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      const { rerender } = render(<WelcomeWizardTrigger />);
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();

      // The device is deleted, dropping the namespace back to zero. The wizard
      // must NOT pop back open in the user's face — only a fresh page load
      // reconsiders eligibility.
      mockUseStats.mockReturnValue({
        stats: zeroStats,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });
      rerender(<WelcomeWizardTrigger />);

      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
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
