import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { mockStats } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getStatusDevices: vi.fn(),
  }),
);

vi.mock("@/utils/welcomeState", () => ({
  hasSeenWelcome: vi.fn(),
  markWelcomeSeen: vi.fn(),
}));

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

import { hasSeenWelcome, markWelcomeSeen } from "@/utils/welcomeState";
import WelcomeWizardTrigger from "../WelcomeWizardTrigger";

const mockHasSeenWelcome = vi.mocked(hasSeenWelcome);
const mockMarkWelcomeSeen = vi.mocked(markWelcomeSeen);

function renderTrigger() {
  return render(<WelcomeWizardTrigger />, { wrapper: createTestWrapper() });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getStatusDevices.mockReturnValue(new Promise(() => {}));
  mockHasSeenWelcome.mockReturnValue(false);
});

describe("WelcomeWizardTrigger", () => {
  describe("when tenant has already seen welcome", () => {
    it("renders nothing", async () => {
      mockHasSeenWelcome.mockReturnValue(true);
      sdk.getStatusDevices.mockResolvedValue(mockSdkResponse(mockStats()));
      renderTrigger();
      await waitFor(() => expect(sdk.getStatusDevices).toHaveBeenCalled());
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when tenant has not seen welcome and has zero devices", () => {
    beforeEach(() => {
      sdk.getStatusDevices.mockResolvedValue(mockSdkResponse(mockStats()));
    });

    it("shows the wizard", async () => {
      renderTrigger();
      expect(await screen.findByTestId("welcome-wizard")).toBeInTheDocument();
    });

    it("does not call markWelcomeSeen when merely closed", async () => {
      renderTrigger();
      (await screen.findByTestId("welcome-wizard")).click();
      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();
    });

    it("calls markWelcomeSeen with the tenant id when dismissed for good", async () => {
      renderTrigger();
      (await screen.findByTestId("wizard-dismiss")).click();
      await waitFor(() => {
        expect(mockMarkWelcomeSeen).toHaveBeenCalledWith("tenant-456");
      });
    });

    it("hides the wizard after it is closed", async () => {
      renderTrigger();
      (await screen.findByTestId("welcome-wizard")).click();
      await waitFor(() => {
        expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
      });
    });

    it("refetches stats when wizard is closed", async () => {
      renderTrigger();
      const callsBefore = sdk.getStatusDevices.mock.calls.length;
      (await screen.findByTestId("welcome-wizard")).click();
      await waitFor(() => {
        expect(sdk.getStatusDevices.mock.calls.length).toBeGreaterThan(
          callsBefore,
        );
      });
    });
  });

  describe("when tenant has devices", () => {
    it("does not show the wizard when there are registered devices", async () => {
      sdk.getStatusDevices.mockResolvedValue(
        mockSdkResponse(mockStats({ registered_devices: 1 })),
      );
      renderTrigger();
      await waitFor(() => expect(sdk.getStatusDevices).toHaveBeenCalled());
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });

    it("still shows the wizard when a device is only pending (not accepted)", async () => {
      sdk.getStatusDevices.mockResolvedValue(
        mockSdkResponse(mockStats({ pending_devices: 2 })),
      );
      renderTrigger();
      expect(await screen.findByTestId("welcome-wizard")).toBeInTheDocument();
    });

    it("still shows the wizard when a device is only rejected (not accepted)", async () => {
      sdk.getStatusDevices.mockResolvedValue(
        mockSdkResponse(mockStats({ rejected_devices: 1 })),
      );
      renderTrigger();
      expect(await screen.findByTestId("welcome-wizard")).toBeInTheDocument();
    });

    it("does not call markWelcomeSeen when there are devices", async () => {
      sdk.getStatusDevices.mockResolvedValue(
        mockSdkResponse(mockStats({ registered_devices: 5 })),
      );
      renderTrigger();
      await waitFor(() => expect(sdk.getStatusDevices).toHaveBeenCalled());
      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();
    });
  });

  describe("eligibility is decided once, at page load", () => {
    it("does not reopen when the last device is deleted mid-session", async () => {
      sdk.getStatusDevices.mockResolvedValue(
        mockSdkResponse(mockStats({ registered_devices: 1 })),
      );
      const { rerender } = renderTrigger();
      await waitFor(() => expect(sdk.getStatusDevices).toHaveBeenCalled());
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();

      sdk.getStatusDevices.mockResolvedValue(mockSdkResponse(mockStats()));
      rerender(<WelcomeWizardTrigger />);
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when stats are loading", () => {
    it("does not show the wizard", () => {
      renderTrigger();
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when tenant is null", () => {
    it("does not show the wizard", async () => {
      seedAuthStore({ tenant: null });
      sdk.getStatusDevices.mockResolvedValue(mockSdkResponse(mockStats()));
      renderTrigger();
      await waitFor(() => expect(sdk.getStatusDevices).toHaveBeenCalled());
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });
});
