import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

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

// Mock the step sub-components to keep tests focused on the orchestrator. The
// install and code faces are presentational markers; the link-path watcher
// (mounted across both faces of step 1) exposes onConnected so a test can
// simulate the device being accepted via the link from either face.
vi.mock("../WizardStepInstall", () => ({
  default: () => <div data-testid="step-install">Install step</div>,
}));

vi.mock("../WizardStepCode", () => ({
  default: () => <div data-testid="step-code">Code step</div>,
}));

vi.mock("../WizardAcceptedWatcher", () => ({
  default: ({
    onConnected,
  }: {
    onConnected: (d: { uid: string; name: string }) => void;
  }) => (
    <button
      type="button"
      onClick={() => onConnected({ uid: "dev-uid-123", name: "my-device" })}
      data-testid="simulate-connected"
    >
      Simulate device connected
    </button>
  ),
}));

vi.mock("../WizardStepComplete", () => ({
  default: () => <div data-testid="step-complete">Complete content</div>,
}));

// The code face's Accept button is footer-driven off these two hooks (owned by
// the orchestrator). Stub them: a complete code and a submit that accepts.
vi.mock("@/hooks/useOtpInput", () => ({
  useOtpInput: () => ({
    code: Array(8).fill("A"),
    inputRefs: { current: [] },
    handleChange: vi.fn(),
    handleKeyDown: vi.fn(),
    handlePaste: vi.fn(),
    reset: vi.fn(),
    getValue: () => "AAAAAAAA",
    isComplete: true,
  }),
}));

vi.mock("@/hooks/useAcceptDeviceByCode", () => ({
  useAcceptDeviceByCode: () => ({
    submit: vi.fn().mockResolvedValue({ uid: "code-uid", name: "code-device" }),
    isPending: false,
    error: "",
    clearError: vi.fn(),
  }),
}));

import WelcomeWizard from "../WelcomeWizard";

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(cleanup);

function renderWizard(open = true, onClose = vi.fn(), onDismiss = vi.fn()) {
  return {
    onClose,
    onDismiss,
    ...render(
      <MemoryRouter>
        <WelcomeWizard open={open} onClose={onClose} onDismiss={onDismiss} />
      </MemoryRouter>,
    ),
  };
}

describe("WelcomeWizard", () => {
  describe("when open=false", () => {
    it("renders nothing", () => {
      render(
        <MemoryRouter>
          <WelcomeWizard open={false} onClose={vi.fn()} onDismiss={vi.fn()} />
        </MemoryRouter>,
      );
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

  describe("Step 1 (install)", () => {
    it("opens on the install step, at step 1 of 2", () => {
      renderWizard();
      expect(
        screen.getByRole("progressbar", { name: /step 1 of 2/i }),
      ).toBeInTheDocument();
      expect(screen.getByTestId("step-install")).toBeInTheDocument();
    });

    it("has an enabled 'Next' and a 'Skip'", () => {
      renderWizard();
      expect(screen.getByRole("button", { name: /next/i })).toBeEnabled();
      expect(
        screen.getByRole("button", { name: /^skip$/i }),
      ).toBeInTheDocument();
    });

    it("clicking 'Skip' dismisses for good, not just closes", async () => {
      const user = userEvent.setup();
      const { onClose, onDismiss } = renderWizard();

      await user.click(screen.getByRole("button", { name: /^skip$/i }));

      expect(onDismiss).toHaveBeenCalledOnce();
      expect(onClose).not.toHaveBeenCalled();
    });

    it("clicking the X (aria-label 'Close wizard') defers via onClose", async () => {
      const user = userEvent.setup();
      const { onClose } = renderWizard();

      await user.click(screen.getByRole("button", { name: /close wizard/i }));

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("auto-advances to the final step when the device connects via the link", async () => {
      const user = userEvent.setup();
      renderWizard();

      await user.click(screen.getByTestId("simulate-connected"));

      expect(
        screen.getByRole("progressbar", { name: /step 2 of 2/i }),
      ).toBeInTheDocument();
      expect(screen.getByTestId("step-complete")).toBeInTheDocument();
    });
  });

  describe("Code-entry face (the 'not on that machine' path)", () => {
    it("clicking 'Next' switches to the code step, still at step 1 of 2", async () => {
      const user = userEvent.setup();
      renderWizard();

      await user.click(screen.getByRole("button", { name: /next/i }));

      expect(screen.getByTestId("step-code")).toBeInTheDocument();
      expect(screen.queryByTestId("step-install")).not.toBeInTheDocument();
      expect(
        screen.getByRole("progressbar", { name: /step 1 of 2/i }),
      ).toBeInTheDocument();
    });

    it("the footer 'Back' returns to the install step", async () => {
      const user = userEvent.setup();
      renderWizard();

      await user.click(screen.getByRole("button", { name: /next/i }));
      await user.click(screen.getByRole("button", { name: /^back$/i }));

      expect(screen.getByTestId("step-install")).toBeInTheDocument();
      expect(screen.queryByTestId("step-code")).not.toBeInTheDocument();
    });

    it("the footer 'Pair device' advances to the final step", async () => {
      const user = userEvent.setup();
      renderWizard();

      await user.click(screen.getByRole("button", { name: /next/i }));
      await user.click(screen.getByRole("button", { name: /pair device/i }));

      expect(
        await screen.findByRole("progressbar", { name: /step 2 of 2/i }),
      ).toBeInTheDocument();
      expect(screen.getByTestId("step-complete")).toBeInTheDocument();
    });

    it("accepting via the link while on the code face still advances", async () => {
      const user = userEvent.setup();
      renderWizard();

      await user.click(screen.getByRole("button", { name: /next/i }));
      // The watcher stays mounted on the code face, so a link acceptance here
      // must still drive the wizard forward.
      await user.click(screen.getByTestId("simulate-connected"));

      expect(screen.getByTestId("step-complete")).toBeInTheDocument();
    });
  });

  describe("Final step", () => {
    async function goToFinal() {
      const user = userEvent.setup();
      const result = renderWizard();
      await user.click(screen.getByTestId("simulate-connected"));
      return { user, ...result };
    }

    it("shows 'Finish' and keeps the X close affordance", async () => {
      await goToFinal();
      expect(
        screen.getByRole("button", { name: /finish/i }),
      ).toBeInTheDocument();
      // The X stays available on the final step (onboarding is done, so it
      // dismisses for good rather than deferring).
      expect(
        screen.getByRole("button", { name: /close wizard/i }),
      ).toBeInTheDocument();
    });

    it("clicking 'Finish' dismisses for good", async () => {
      const { user, onClose, onDismiss } = await goToFinal();

      await user.click(screen.getByRole("button", { name: /finish/i }));

      expect(onDismiss).toHaveBeenCalled();
      expect(onClose).not.toHaveBeenCalled();
    });

    it("clicking the X on the final step dismisses for good", async () => {
      const { user, onClose, onDismiss } = await goToFinal();

      await user.click(screen.getByRole("button", { name: /close wizard/i }));

      expect(onDismiss).toHaveBeenCalledOnce();
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("closing routes to defer or dismiss by step", () => {
    async function goToFinal() {
      const user = userEvent.setup();
      const result = renderWizard();
      await user.click(screen.getByTestId("simulate-connected"));
      result.onClose.mockClear();
      result.onDismiss.mockClear();
      return result;
    }

    it("Escape defers (onClose) on step 1 but dismisses (onDismiss) on the final step", async () => {
      const { onClose, onDismiss } = renderWizard();
      fireEvent(screen.getByRole("dialog"), new Event("cancel"));
      expect(onClose).toHaveBeenCalled();
      expect(onDismiss).not.toHaveBeenCalled();
      cleanup();

      const final = await goToFinal();
      fireEvent(screen.getByRole("dialog"), new Event("cancel"));
      expect(final.onDismiss).toHaveBeenCalled();
      expect(final.onClose).not.toHaveBeenCalled();
    });

    it("backdrop click defers on step 1 but dismisses on the final step", async () => {
      const { onClose, onDismiss } = renderWizard();
      let dialog = document.querySelector("dialog") as HTMLElement;
      fireEvent.mouseDown(dialog);
      fireEvent.click(dialog);
      expect(onClose).toHaveBeenCalled();
      expect(onDismiss).not.toHaveBeenCalled();
      cleanup();

      const final = await goToFinal();
      dialog = document.querySelector("dialog") as HTMLElement;
      fireEvent.mouseDown(dialog);
      fireEvent.click(dialog);
      expect(final.onDismiss).toHaveBeenCalled();
      expect(final.onClose).not.toHaveBeenCalled();
    });
  });
});
