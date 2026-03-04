import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useDevicesStore } from "@/stores/devicesStore";

// Mock the focus trap so it doesn't interfere with jsdom focus state
vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));

// jsdom doesn't implement showModal/close — stub them so they behave like the
// open attribute (which testing-library uses to resolve the dialog's role)
HTMLDialogElement.prototype.showModal = vi.fn(function(this: HTMLDialogElement) {
  this.setAttribute("open", "");
});
HTMLDialogElement.prototype.close = vi.fn(function(this: HTMLDialogElement) {
  this.removeAttribute("open");
});

// Mock all four step sub-components to keep tests focused on the orchestrator
vi.mock("../WizardStep1Welcome", () => ({
  default: () => <div data-testid="step-1-welcome">Step 1 content</div>,
}));

vi.mock("../WizardStep2Install", () => ({
  default: ({ onDeviceDetected }: { onDeviceDetected: () => void }) => (
    <div data-testid="step-2-install">
      Step 2 content
      <button onClick={onDeviceDetected} data-testid="simulate-device-detected">
        Simulate device detected
      </button>
    </div>
  ),
}));

vi.mock("../WizardStep3Device", () => ({
  default: ({
    onDeviceLoaded,
  }: {
    device: null;
    onDeviceLoaded: (d: { uid: string; name: string }) => void;
  }) => (
    <div data-testid="step-3-device">
      Step 3 content
      <button
        onClick={() => onDeviceLoaded({ uid: "dev-uid-123", name: "my-device" } as never)}
        data-testid="simulate-device-loaded"
      >
        Load device
      </button>
    </div>
  ),
}));

vi.mock("../WizardStep4Complete", () => ({
  default: () => <div data-testid="step-4-complete">Step 4 content</div>,
}));

import WelcomeWizard from "../WelcomeWizard";

const mockAccept = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  useDevicesStore.setState({ accept: mockAccept } as never);
});

afterEach(cleanup);

function renderWizard(open = true, onClose = vi.fn()) {
  return { onClose, ...render(<WelcomeWizard open={open} onClose={onClose} />) };
}

describe("WelcomeWizard", () => {
  describe("when open=false", () => {
    it("renders nothing", () => {
      render(<WelcomeWizard open={false} onClose={vi.fn()} />);
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("when open=true", () => {
    it("renders the dialog", () => {
      renderWizard();
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    it("dialog has the accessible label", () => {
      renderWizard();
      expect(
        screen.getByRole("dialog", { name: /welcome to shellhub/i }),
      ).toBeInTheDocument();
    });
  });

  describe("Step 1", () => {
    it("shows step indicator at step 1", () => {
      renderWizard();
      expect(screen.getByRole("progressbar", { name: /step 1 of 4/i })).toBeInTheDocument();
    });

    it("renders step 1 content", () => {
      renderWizard();
      expect(screen.getByTestId("step-1-welcome")).toBeInTheDocument();
    });

    it("has a 'Next' button", () => {
      renderWizard();
      expect(screen.getByRole("button", { name: /next/i })).toBeInTheDocument();
    });

    it("has a 'Close' button", () => {
      renderWizard();
      // The visible "Close" text button in the footer
      expect(screen.getByRole("button", { name: /^close$/i })).toBeInTheDocument();
    });

    it("clicking 'Next' advances to step 2", async () => {
      const user = userEvent.setup();
      renderWizard();

      await user.click(screen.getByRole("button", { name: /next/i }));

      expect(screen.getByRole("progressbar", { name: /step 2 of 4/i })).toBeInTheDocument();
    });

    it("clicking 'Close' button calls onClose", async () => {
      const user = userEvent.setup();
      const { onClose } = renderWizard();

      await user.click(screen.getByRole("button", { name: /^close$/i }));

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("clicking the X (aria-label 'Close wizard') calls onClose", async () => {
      const user = userEvent.setup();
      const { onClose } = renderWizard();

      await user.click(screen.getByRole("button", { name: /close wizard/i }));

      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe("Step 2", () => {
    async function goToStep2() {
      const user = userEvent.setup();
      const result = renderWizard();
      await user.click(screen.getByRole("button", { name: /next/i }));
      return { user, ...result };
    }

    it("shows step indicator at step 2", async () => {
      await goToStep2();
      expect(screen.getByRole("progressbar", { name: /step 2 of 4/i })).toBeInTheDocument();
    });

    it("renders step 2 content", async () => {
      await goToStep2();
      expect(screen.getByTestId("step-2-install")).toBeInTheDocument();
    });

    it("'Next' button is present but disabled", async () => {
      await goToStep2();
      const nextBtn = screen.getByRole("button", { name: /next/i });
      expect(nextBtn).toBeDisabled();
    });

    it("'Close' button is present", async () => {
      await goToStep2();
      expect(screen.getByRole("button", { name: /^close$/i })).toBeInTheDocument();
    });

    it("auto-advances to step 3 when onDeviceDetected is called", async () => {
      const { user } = await goToStep2();

      await user.click(screen.getByTestId("simulate-device-detected"));

      expect(screen.getByRole("progressbar", { name: /step 3 of 4/i })).toBeInTheDocument();
    });
  });

  describe("Step 3", () => {
    async function goToStep3() {
      const user = userEvent.setup();
      const result = renderWizard();
      // Step 1 → 2
      await user.click(screen.getByRole("button", { name: /next/i }));
      // Simulate device detected → step 3
      await user.click(screen.getByTestId("simulate-device-detected"));
      return { user, ...result };
    }

    it("shows step indicator at step 3", async () => {
      await goToStep3();
      expect(screen.getByRole("progressbar", { name: /step 3 of 4/i })).toBeInTheDocument();
    });

    it("renders step 3 content", async () => {
      await goToStep3();
      expect(screen.getByTestId("step-3-device")).toBeInTheDocument();
    });

    it("has an 'Accept' button and a 'Close' button", async () => {
      await goToStep3();
      expect(screen.getByRole("button", { name: /^accept$/i })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /^close$/i })).toBeInTheDocument();
    });

    it("'Accept' is disabled when no device is loaded", async () => {
      await goToStep3();
      // No device loaded yet — button should be disabled
      expect(screen.getByRole("button", { name: /^accept$/i })).toBeDisabled();
    });

    it("clicking 'Accept' after loading a device calls accept(uid) and advances to step 4", async () => {
      mockAccept.mockResolvedValue(undefined);
      const { user } = await goToStep3();

      // Load a device first
      await user.click(screen.getByTestId("simulate-device-loaded"));

      await user.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => {
        expect(mockAccept).toHaveBeenCalledWith("dev-uid-123");
        expect(screen.getByRole("progressbar", { name: /step 4 of 4/i })).toBeInTheDocument();
      });
    });

    it("stays on step 3 when accept() throws", async () => {
      mockAccept.mockRejectedValue(new Error("API error"));
      const { user } = await goToStep3();

      await user.click(screen.getByTestId("simulate-device-loaded"));
      await user.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => {
        expect(screen.getByRole("progressbar", { name: /step 3 of 4/i })).toBeInTheDocument();
      });
    });

    it("clicking 'Close' calls onClose without touching the device", async () => {
      const { user, onClose } = await goToStep3();

      await user.click(screen.getByRole("button", { name: /^close$/i }));

      expect(onClose).toHaveBeenCalledOnce();
      expect(mockAccept).not.toHaveBeenCalled();
    });
  });

  describe("Step 4", () => {
    async function goToStep4() {
      mockAccept.mockResolvedValue(undefined);
      const user = userEvent.setup();
      const result = renderWizard();
      // 1 → 2
      await user.click(screen.getByRole("button", { name: /next/i }));
      // 2 → 3 (device detected)
      await user.click(screen.getByTestId("simulate-device-detected"));
      // Load device
      await user.click(screen.getByTestId("simulate-device-loaded"));
      // Accept → 4
      await user.click(screen.getByRole("button", { name: /^accept$/i }));
      await waitFor(() => expect(screen.getByRole("progressbar", { name: /step 4 of 4/i })).toBeInTheDocument());
      return { user, ...result };
    }

    it("shows step indicator at step 4", async () => {
      await goToStep4();
      expect(screen.getByRole("progressbar", { name: /step 4 of 4/i })).toBeInTheDocument();
    });

    it("renders step 4 content", async () => {
      await goToStep4();
      expect(screen.getByTestId("step-4-complete")).toBeInTheDocument();
    });

    it("has a 'Finish' button and no 'Close' button", async () => {
      await goToStep4();
      expect(screen.getByRole("button", { name: /finish/i })).toBeInTheDocument();
      // The text "Close" button should not exist on step 4
      expect(screen.queryByRole("button", { name: /^close$/i })).not.toBeInTheDocument();
    });

    it("X (close wizard) button is hidden on step 4", async () => {
      await goToStep4();
      expect(
        screen.queryByRole("button", { name: /close wizard/i }),
      ).not.toBeInTheDocument();
    });

    it("clicking 'Finish' calls onClose", async () => {
      const { user, onClose } = await goToStep4();

      await user.click(screen.getByRole("button", { name: /finish/i }));

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe("progress bar", () => {
    it("starts at 25% width on step 1", () => {
      renderWizard();
      const bar = document.querySelector("[style*='width']") as HTMLElement;
      expect(bar?.style.width).toBe("25%");
    });

    it("advances to 50% width on step 2", async () => {
      const user = userEvent.setup();
      renderWizard();

      await user.click(screen.getByRole("button", { name: /next/i }));

      const bar = document.querySelector("[style*='width']") as HTMLElement;
      expect(bar?.style.width).toBe("50%");
    });
  });

  describe("Escape key", () => {
    it("calls onClose when Escape is pressed on step 1", async () => {
      const { onClose } = renderWizard();

      // showModal dialogs fire a native "cancel" event on Escape
      fireEvent(screen.getByRole("dialog"), new Event("cancel"));

      expect(onClose).toHaveBeenCalled();
    });

    it("calls onClose when Escape is pressed on step 3", async () => {
      const user = userEvent.setup();
      const { onClose } = renderWizard();

      // Navigate to step 3
      await user.click(screen.getByRole("button", { name: /next/i }));
      await user.click(screen.getByTestId("simulate-device-detected"));

      fireEvent(screen.getByRole("dialog"), new Event("cancel"));

      expect(onClose).toHaveBeenCalled();
    });

    it("does NOT call onClose when Escape is pressed on step 4", async () => {
      mockAccept.mockResolvedValue(undefined);
      const user = userEvent.setup();
      const { onClose } = renderWizard();

      // Navigate to step 4
      await user.click(screen.getByRole("button", { name: /next/i }));
      await user.click(screen.getByTestId("simulate-device-detected"));
      await user.click(screen.getByTestId("simulate-device-loaded"));
      await user.click(screen.getByRole("button", { name: /^accept$/i }));
      await waitFor(() => expect(screen.getByRole("progressbar", { name: /step 4 of 4/i })).toBeInTheDocument());

      onClose.mockClear();
      fireEvent(screen.getByRole("dialog"), new Event("cancel"));

      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("backdrop click", () => {
    it("calls onClose when backdrop is clicked on step 1", async () => {
      const user = userEvent.setup();
      const { onClose } = renderWizard();

      // The backdrop is the absolutely-positioned sibling of the dialog
      const backdrop = document.querySelector("[aria-hidden='true']") as HTMLElement;
      await user.click(backdrop);

      expect(onClose).toHaveBeenCalled();
    });

    it("does NOT call onClose when backdrop is clicked on step 4", async () => {
      mockAccept.mockResolvedValue(undefined);
      const user = userEvent.setup();
      const { onClose } = renderWizard();

      // Navigate to step 4
      await user.click(screen.getByRole("button", { name: /next/i }));
      await user.click(screen.getByTestId("simulate-device-detected"));
      await user.click(screen.getByTestId("simulate-device-loaded"));
      await user.click(screen.getByRole("button", { name: /^accept$/i }));
      await waitFor(() => expect(screen.getByRole("progressbar", { name: /step 4 of 4/i })).toBeInTheDocument());

      onClose.mockClear();

      const backdrop = document.querySelector("[aria-hidden='true']") as HTMLElement;
      await user.click(backdrop);

      expect(onClose).not.toHaveBeenCalled();
    });
  });
});
