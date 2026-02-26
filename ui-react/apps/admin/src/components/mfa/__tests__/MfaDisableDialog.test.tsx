import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import MfaDisableDialog from "../MfaDisableDialog";

vi.mock("../../../api/mfa", () => ({
  disableMfa: vi.fn(),
}));

import { disableMfa } from "../../../api/mfa";

const mockedDisableMfa = vi.mocked(disableMfa);

describe("MfaDisableDialog", () => {
  const onClose = vi.fn();
  const onSuccess = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockedDisableMfa.mockResolvedValue(undefined);
  });

  describe("Mode Switching", () => {
    it("defaults to TOTP mode", () => {
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      expect(screen.getByText(/enter your 6-digit code/i)).toBeInTheDocument();
      expect(screen.getByText(/authenticator app/i)).toBeInTheDocument();
    });

    it("switches to recovery code mode", async () => {
      const user = userEvent.setup();
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      const switchButton = screen.getByText(/use recovery code/i);
      await user.click(switchButton);

      expect(screen.getByText(/enter a recovery code/i)).toBeInTheDocument();
      expect(screen.getByPlaceholderText(/recovery code/i)).toBeInTheDocument();
    });

    it("switches back to TOTP mode from recovery", async () => {
      const user = userEvent.setup();
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Switch to recovery mode
      const switchToRecovery = screen.getByText(/use recovery code/i);
      await user.click(switchToRecovery);

      // Switch back to TOTP
      const switchToTotp = screen.getByText(/use authenticator/i);
      await user.click(switchToTotp);

      expect(screen.getByText(/enter your 6-digit code/i)).toBeInTheDocument();
    });
  });

  describe("TOTP Mode Validation", () => {
    it("requires all 6 digits before enabling submit", async () => {
      const user = userEvent.setup();
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      const disableButton = screen.getByRole("button", { name: /disable mfa/i });
      expect(disableButton).toBeDisabled();

      // Type 6 digits
      const inputs = screen.getAllByRole("textbox");
      const otpInputs = inputs.filter((input) =>
        input.getAttribute("maxLength") === "1"
      );

      await user.type(otpInputs[0], "1");
      await user.type(otpInputs[1], "2");
      await user.type(otpInputs[2], "3");
      await user.type(otpInputs[3], "4");
      await user.type(otpInputs[4], "5");
      await user.type(otpInputs[5], "6");

      expect(disableButton).toBeEnabled();
    });

    it("successfully disables MFA with valid TOTP", async () => {
      const user = userEvent.setup();
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Enter 6-digit code
      const inputs = screen.getAllByRole("textbox");
      const otpInputs = inputs.filter((input) =>
        input.getAttribute("maxLength") === "1"
      );

      await user.type(otpInputs[0], "1");
      await user.type(otpInputs[1], "2");
      await user.type(otpInputs[2], "3");
      await user.type(otpInputs[3], "4");
      await user.type(otpInputs[4], "5");
      await user.type(otpInputs[5], "6");

      const disableButton = screen.getByRole("button", { name: /disable mfa/i });
      await user.click(disableButton);

      await waitFor(() => {
        expect(mockedDisableMfa).toHaveBeenCalledWith({ code: "123456" });
        expect(onSuccess).toHaveBeenCalled();
        expect(onClose).toHaveBeenCalled();
      });
    });

    it("shows error on invalid TOTP", async () => {
      const user = userEvent.setup();
      mockedDisableMfa.mockRejectedValue(new Error("Invalid code"));

      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Enter 6-digit code
      const inputs = screen.getAllByRole("textbox");
      const otpInputs = inputs.filter((input) =>
        input.getAttribute("maxLength") === "1"
      );

      await user.type(otpInputs[0], "9");
      await user.type(otpInputs[1], "9");
      await user.type(otpInputs[2], "9");
      await user.type(otpInputs[3], "9");
      await user.type(otpInputs[4], "9");
      await user.type(otpInputs[5], "9");

      const disableButton = screen.getByRole("button", { name: /disable mfa/i });
      await user.click(disableButton);

      await waitFor(() => {
        expect(screen.getByText(/invalid code/i)).toBeInTheDocument();
      });

      // Should not call onSuccess
      expect(onSuccess).not.toHaveBeenCalled();
    });
  });

  describe("Recovery Code Mode Validation", () => {
    it("requires recovery code before enabling submit", async () => {
      const user = userEvent.setup();
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Switch to recovery mode
      const switchButton = screen.getByText(/use recovery code/i);
      await user.click(switchButton);

      const disableButton = screen.getByRole("button", { name: /disable mfa/i });
      expect(disableButton).toBeDisabled();

      // Type recovery code
      const recoveryInput = screen.getByPlaceholderText(/recovery code/i);
      await user.type(recoveryInput, "abc123xyz");

      expect(disableButton).toBeEnabled();
    });

    it("successfully disables MFA with valid recovery code", async () => {
      const user = userEvent.setup();
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Switch to recovery mode
      const switchButton = screen.getByText(/use recovery code/i);
      await user.click(switchButton);

      // Enter recovery code
      const recoveryInput = screen.getByPlaceholderText(/recovery code/i);
      await user.type(recoveryInput, "valid-recovery-code");

      const disableButton = screen.getByRole("button", { name: /disable mfa/i });
      await user.click(disableButton);

      await waitFor(() => {
        expect(mockedDisableMfa).toHaveBeenCalledWith({
          recovery_code: "valid-recovery-code",
        });
        expect(onSuccess).toHaveBeenCalled();
        expect(onClose).toHaveBeenCalled();
      });
    });

    it("shows error on invalid recovery code", async () => {
      const user = userEvent.setup();
      mockedDisableMfa.mockRejectedValue(new Error("Invalid recovery code"));

      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Switch to recovery mode
      const switchButton = screen.getByText(/use recovery code/i);
      await user.click(switchButton);

      // Enter recovery code
      const recoveryInput = screen.getByPlaceholderText(/recovery code/i);
      await user.type(recoveryInput, "invalid-code");

      const disableButton = screen.getByRole("button", { name: /disable mfa/i });
      await user.click(disableButton);

      await waitFor(() => {
        expect(screen.getByText(/invalid code/i)).toBeInTheDocument();
      });

      expect(onSuccess).not.toHaveBeenCalled();
    });
  });

  describe("Dialog Behavior", () => {
    it("closes when cancel button is clicked", async () => {
      const user = userEvent.setup();
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      const cancelButton = screen.getByText(/cancel/i);
      await user.click(cancelButton);

      expect(onClose).toHaveBeenCalled();
      expect(onSuccess).not.toHaveBeenCalled();
    });

    it("closes when clicking outside (backdrop)", async () => {
      const user = userEvent.setup();
      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Get the backdrop (parent div with onClick)
      const backdrop = screen.getByText(/disable two-factor/i).parentElement?.parentElement
        ?.previousElementSibling;

      if (backdrop) {
        await user.click(backdrop);
        expect(onClose).toHaveBeenCalled();
      }
    });

    it("does not render when open is false", () => {
      const { container } = render(
        <MfaDisableDialog open={false} onClose={onClose} onSuccess={onSuccess} />
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe("Loading State", () => {
    it("shows loading state while submitting", async () => {
      const user = userEvent.setup();
      let resolveDisable: () => void;
      mockedDisableMfa.mockReturnValue(
        new Promise((resolve) => {
          resolveDisable = resolve as () => void;
        })
      );

      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Enter code
      const inputs = screen.getAllByRole("textbox");
      const otpInputs = inputs.filter((input) =>
        input.getAttribute("maxLength") === "1"
      );

      await user.type(otpInputs[0], "1");
      await user.type(otpInputs[1], "2");
      await user.type(otpInputs[2], "3");
      await user.type(otpInputs[3], "4");
      await user.type(otpInputs[4], "5");
      await user.type(otpInputs[5], "6");

      const disableButton = screen.getByRole("button", { name: /disable mfa/i });
      await user.click(disableButton);

      // Should show loading state
      expect(screen.getByText(/disabling/i)).toBeInTheDocument();
      expect(disableButton).toBeDisabled();

      // Resolve the promise
      resolveDisable!();

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalled();
      });
    });
  });

  describe("Error Handling", () => {
    it("clears previous errors when switching modes", async () => {
      const user = userEvent.setup();
      mockedDisableMfa.mockRejectedValue(new Error("Invalid code"));

      render(
        <MfaDisableDialog open={true} onClose={onClose} onSuccess={onSuccess} />
      );

      // Try with invalid TOTP
      const inputs = screen.getAllByRole("textbox");
      const otpInputs = inputs.filter((input) =>
        input.getAttribute("maxLength") === "1"
      );

      await user.type(otpInputs[0], "9");
      await user.type(otpInputs[1], "9");
      await user.type(otpInputs[2], "9");
      await user.type(otpInputs[3], "9");
      await user.type(otpInputs[4], "9");
      await user.type(otpInputs[5], "9");

      const disableButton = screen.getByRole("button", { name: /disable mfa/i });
      await user.click(disableButton);

      await waitFor(() => {
        expect(screen.getByText(/invalid code/i)).toBeInTheDocument();
      });

      // Switch to recovery mode
      const switchButton = screen.getByText(/use recovery code/i);
      await user.click(switchButton);

      // Error should be cleared
      expect(screen.queryByText(/invalid code/i)).not.toBeInTheDocument();
    });
  });
});
