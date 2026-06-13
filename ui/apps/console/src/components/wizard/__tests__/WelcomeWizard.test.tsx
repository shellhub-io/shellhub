import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// Mock the focus trap so it doesn't interfere with jsdom focus state
vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));

// jsdom doesn't implement showModal/close — stub them so they behave like the
// open attribute (which testing-library uses to resolve the dialog's role)
HTMLDialogElement.prototype.showModal = vi.fn(function (
  this: HTMLDialogElement,
) {
  this.setAttribute("open", "");
});
HTMLDialogElement.prototype.close = vi.fn(function (this: HTMLDialogElement) {
  this.removeAttribute("open");
});

// Mock the step sub-components to keep tests focused on the orchestrator. Step 2
// exposes onConnected so a test can simulate the device connecting.
vi.mock("../WizardStep1Welcome", () => ({
  default: () => <div data-testid="step-1-welcome">Step 1 content</div>,
}));

vi.mock("../WizardStep2Install", () => ({
  default: ({
    onConnected,
  }: {
    onConnected: (d: { uid: string; name: string }) => void;
  }) => (
    <div data-testid="step-2-install">
      Step 2 content
      <button
        type="button"
        onClick={() => onConnected({ uid: "dev-uid-123", name: "my-device" })}
        data-testid="simulate-connected"
      >
        Simulate device connected
      </button>
    </div>
  ),
}));

vi.mock("../WizardStep4Complete", () => ({
  default: () => <div data-testid="step-complete">Complete content</div>,
}));

import WelcomeWizard from "../WelcomeWizard";

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(cleanup);

function renderWizard(open = true, onClose = vi.fn()) {
  return {
    onClose,
    ...render(<WelcomeWizard open={open} onClose={onClose} />),
  };
}

describe("WelcomeWizard", () => {
  describe("when open=false", () => {
    it("renders nothing", () => {
      render(<WelcomeWizard open={false} onClose={vi.fn()} />);
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("when open=true", () => {
    it("renders the dialog with the accessible label", () => {
      renderWizard();
      expect(
        screen.getByRole("dialog", { name: /welcome to shellhub/i }),
      ).toBeInTheDocument();
    });
  });

  describe("Step 1", () => {
    it("shows step indicator at step 1 of 3 and renders its content", () => {
      renderWizard();
      expect(
        screen.getByRole("progressbar", { name: /step 1 of 3/i }),
      ).toBeInTheDocument();
      expect(screen.getByTestId("step-1-welcome")).toBeInTheDocument();
    });

    it("clicking 'Next' advances to step 2", async () => {
      const user = userEvent.setup();
      renderWizard();

      await user.click(screen.getByRole("button", { name: /next/i }));

      expect(
        screen.getByRole("progressbar", { name: /step 2 of 3/i }),
      ).toBeInTheDocument();
    });

    it("clicking 'Close' calls onClose", async () => {
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

    it("renders step 2 with a disabled 'Next' and a 'Close'", async () => {
      await goToStep2();
      expect(screen.getByTestId("step-2-install")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /next/i })).toBeDisabled();
      expect(
        screen.getByRole("button", { name: /^close$/i }),
      ).toBeInTheDocument();
    });

    it("auto-advances to the final step when the device connects", async () => {
      const { user } = await goToStep2();

      await user.click(screen.getByTestId("simulate-connected"));

      expect(
        screen.getByRole("progressbar", { name: /step 3 of 3/i }),
      ).toBeInTheDocument();
      expect(screen.getByTestId("step-complete")).toBeInTheDocument();
    });
  });

  describe("Final step", () => {
    async function goToFinal() {
      const user = userEvent.setup();
      const result = renderWizard();
      await user.click(screen.getByRole("button", { name: /next/i }));
      await user.click(screen.getByTestId("simulate-connected"));
      return { user, ...result };
    }

    it("shows 'Finish' and hides the close affordances", async () => {
      await goToFinal();
      expect(
        screen.getByRole("button", { name: /finish/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /^close$/i }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /close wizard/i }),
      ).not.toBeInTheDocument();
    });

    it("clicking 'Finish' calls onClose", async () => {
      const { user, onClose } = await goToFinal();

      await user.click(screen.getByRole("button", { name: /finish/i }));

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe("dismissal is blocked on the final step", () => {
    async function goToFinal() {
      const user = userEvent.setup();
      const { onClose } = renderWizard();
      await user.click(screen.getByRole("button", { name: /next/i }));
      await user.click(screen.getByTestId("simulate-connected"));
      onClose.mockClear();
      return onClose;
    }

    it("Escape calls onClose on step 1 but not on the final step", async () => {
      const { onClose } = renderWizard();
      fireEvent(screen.getByRole("dialog"), new Event("cancel"));
      expect(onClose).toHaveBeenCalled();
      cleanup();

      const finalOnClose = await goToFinal();
      fireEvent(screen.getByRole("dialog"), new Event("cancel"));
      expect(finalOnClose).not.toHaveBeenCalled();
    });

    it("backdrop click calls onClose on step 1 but not on the final step", async () => {
      const { onClose } = renderWizard();
      let dialog = document.querySelector("dialog") as HTMLElement;
      fireEvent.mouseDown(dialog);
      fireEvent.click(dialog);
      expect(onClose).toHaveBeenCalled();
      cleanup();

      const finalOnClose = await goToFinal();
      dialog = document.querySelector("dialog") as HTMLElement;
      fireEvent.mouseDown(dialog);
      fireEvent.click(dialog);
      expect(finalOnClose).not.toHaveBeenCalled();
    });
  });

  describe("progress bar", () => {
    it("advances width across steps", async () => {
      const user = userEvent.setup();
      renderWizard();

      let bar = document.querySelector("[style*='width']") as HTMLElement;
      expect(bar.style.width).toMatch(/^33\./);

      await user.click(screen.getByRole("button", { name: /next/i }));
      bar = document.querySelector("[style*='width']") as HTMLElement;
      expect(bar.style.width).toMatch(/^66\./);
    });
  });
});
