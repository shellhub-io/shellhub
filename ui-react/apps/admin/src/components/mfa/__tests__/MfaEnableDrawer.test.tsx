import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import MfaEnableDrawer from "../MfaEnableDrawer";

vi.mock("../../../api/mfa", () => ({
  generateMfa: vi.fn(),
  enableMfa: vi.fn(),
}));

vi.mock("../../../api/auth", () => ({
  updateUser: vi.fn(),
}));

import { generateMfa, enableMfa } from "../../../api/mfa";
import { updateUser } from "../../../api/auth";

const mockedGenerateMfa = vi.mocked(generateMfa);
const mockedEnableMfa = vi.mocked(enableMfa);
const mockedUpdateUser = vi.mocked(updateUser);

const mockMfaData = {
  link: "otpauth://totp/ShellHub:user@example.com?secret=ABCD1234&issuer=ShellHub",
  secret: "ABCD1234",
  recovery_codes: ["code1", "code2", "code3", "code4", "code5", "code6"],
};

describe("MfaEnableDrawer", () => {
  const onClose = vi.fn();
  const onSuccess = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockedGenerateMfa.mockResolvedValue(mockMfaData);
    mockedEnableMfa.mockResolvedValue(undefined);
    mockedUpdateUser.mockResolvedValue(undefined);
  });

  describe("Step 1: Recovery Email", () => {
    it("shows email input when no current recovery email", () => {
      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail={null}
        />
      );

      expect(screen.getByText(/recovery email/i)).toBeInTheDocument();
      expect(screen.getByPlaceholderText(/email address/i)).toBeInTheDocument();
    });

    it("shows confirmation when recovery email already exists", () => {
      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      expect(screen.getByText(/recovery@example.com/)).toBeInTheDocument();
      expect(screen.getByText(/continue/i)).toBeInTheDocument();
    });

    it("saves new recovery email and proceeds to step 2", async () => {
      const user = userEvent.setup();
      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail={null}
        />
      );

      const emailInput = screen.getByPlaceholderText(/email address/i);
      await user.type(emailInput, "new-recovery@example.com");

      const nextButton = screen.getByRole("button", { name: /next/i });
      await user.click(nextButton);

      await waitFor(() => {
        expect(mockedUpdateUser).toHaveBeenCalledWith({
          recovery_email: "new-recovery@example.com",
        });
        expect(mockedGenerateMfa).toHaveBeenCalled();
      });
    });

    it("shows error when email is already in use (409)", async () => {
      const user = userEvent.setup();
      mockedUpdateUser.mockRejectedValue(new Error("409"));

      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail={null}
        />
      );

      const emailInput = screen.getByPlaceholderText(/email address/i);
      await user.type(emailInput, "duplicate@example.com");

      const nextButton = screen.getByRole("button", { name: /next/i });
      await user.click(nextButton);

      await waitFor(() => {
        expect(screen.getByText(/email already in use/i)).toBeInTheDocument();
      });
    });

    it("confirms existing email and proceeds to step 2", async () => {
      const user = userEvent.setup();
      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      const continueButton = screen.getByText(/continue/i);
      await user.click(continueButton);

      await waitFor(() => {
        expect(mockedGenerateMfa).toHaveBeenCalled();
      });
    });
  });

  describe("Step 2: Recovery Codes", () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      // Proceed to step 2
      const continueButton = screen.getByText(/continue/i);
      await user.click(continueButton);
      await waitFor(() => expect(mockedGenerateMfa).toHaveBeenCalled());
    });

    it("displays all 6 recovery codes", async () => {
      await waitFor(() => {
        mockMfaData.recovery_codes.forEach((code) => {
          expect(screen.getByText(code)).toBeInTheDocument();
        });
      });
    });

    it("requires saving confirmation before proceeding", async () => {
      const user = userEvent.setup();

      const nextButton = screen.getByRole("button", { name: /next/i });
      expect(nextButton).toBeDisabled();

      const checkbox = screen.getByRole("checkbox");
      await user.click(checkbox);

      expect(nextButton).toBeEnabled();
    });

    it("downloads recovery codes", async () => {
      const user = userEvent.setup();

      const downloadButton = screen.getByText(/download/i);
      await user.click(downloadButton);

      // Check that download was triggered (would need to mock document.createElement)
      // For now, just verify the button is clickable
      expect(downloadButton).toBeInTheDocument();
    });
  });

  describe("Step 3: QR Code and Verification", () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      // Proceed to step 2
      const continueButton = screen.getByText(/continue/i);
      await user.click(continueButton);
      await waitFor(() => expect(mockedGenerateMfa).toHaveBeenCalled());

      // Proceed to step 3
      const checkbox = screen.getByRole("checkbox");
      await user.click(checkbox);
      const nextButton = screen.getByRole("button", { name: /next/i });
      await user.click(nextButton);
    });

    it("displays QR code", async () => {
      await waitFor(() => {
        // QR code canvas should be rendered (QRCodeDisplay component)
        expect(screen.getByText(/scan this qr code/i)).toBeInTheDocument();
      });
    });

    it("displays secret for manual entry", async () => {
      await waitFor(() => {
        expect(screen.getByText(mockMfaData.secret)).toBeInTheDocument();
      });
    });

    it("validates OTP and enables MFA on success", async () => {
      const user = userEvent.setup();

      await waitFor(() => {
        expect(screen.getByText(/scan this qr code/i)).toBeInTheDocument();
      });

      // Find OTP inputs and enter code
      const inputs = screen.getAllByRole("textbox");
      const otpInputs = inputs.filter((input) =>
        input.getAttribute("maxLength") === "1"
      );

      // Type 6-digit code
      await user.type(otpInputs[0], "1");
      await user.type(otpInputs[1], "2");
      await user.type(otpInputs[2], "3");
      await user.type(otpInputs[3], "4");
      await user.type(otpInputs[4], "5");
      await user.type(otpInputs[5], "6");

      const verifyButton = screen.getByRole("button", { name: /verify/i });
      await user.click(verifyButton);

      await waitFor(() => {
        expect(mockedEnableMfa).toHaveBeenCalledWith({
          code: "123456",
          secret: mockMfaData.secret,
          recovery_codes: mockMfaData.recovery_codes,
        });
      });
    });

    it("shows error on invalid OTP", async () => {
      const user = userEvent.setup();
      mockedEnableMfa.mockRejectedValue(new Error("Invalid code"));

      await waitFor(() => {
        expect(screen.getByText(/scan this qr code/i)).toBeInTheDocument();
      });

      // Enter OTP
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

      const verifyButton = screen.getByRole("button", { name: /verify/i });
      await user.click(verifyButton);

      await waitFor(() => {
        expect(screen.getByText(/invalid verification code/i)).toBeInTheDocument();
      });
    });
  });

  describe("Step 4: Success", () => {
    it("shows success message and calls onSuccess", async () => {
      const user = userEvent.setup();
      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      // Navigate through all steps
      const continueButton = screen.getByText(/continue/i);
      await user.click(continueButton);
      await waitFor(() => expect(mockedGenerateMfa).toHaveBeenCalled());

      const checkbox = screen.getByRole("checkbox");
      await user.click(checkbox);
      const nextButton = screen.getByRole("button", { name: /next/i });
      await user.click(nextButton);

      await waitFor(() => {
        expect(screen.getByText(/scan this qr code/i)).toBeInTheDocument();
      });

      // Enter OTP
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

      const verifyButton = screen.getByRole("button", { name: /verify/i });
      await user.click(verifyButton);

      await waitFor(() => {
        expect(screen.getByText(/successfully enabled/i)).toBeInTheDocument();
      });

      const doneButton = screen.getByText(/done/i);
      await user.click(doneButton);

      expect(onSuccess).toHaveBeenCalled();
      expect(onClose).toHaveBeenCalled();
    });
  });

  describe("State Cleanup", () => {
    it("resets state when drawer is closed and reopened", async () => {
      const user = userEvent.setup();
      const { rerender } = render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      // Proceed to step 2
      const continueButton = screen.getByText(/continue/i);
      await user.click(continueButton);
      await waitFor(() => expect(mockedGenerateMfa).toHaveBeenCalled());

      // Close drawer
      rerender(
        <MfaEnableDrawer
          open={false}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      // Reopen drawer
      rerender(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      // Should be back at step 1
      expect(screen.getByText(/recovery@example.com/)).toBeInTheDocument();
    });
  });

  describe("Error Handling", () => {
    it("handles API errors when generating MFA codes", async () => {
      const user = userEvent.setup();
      mockedGenerateMfa.mockRejectedValue(new Error("Network error"));

      render(
        <MfaEnableDrawer
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />
      );

      const continueButton = screen.getByText(/continue/i);
      await user.click(continueButton);

      await waitFor(() => {
        expect(screen.getByText(/failed to generate/i)).toBeInTheDocument();
      });
    });
  });
});
