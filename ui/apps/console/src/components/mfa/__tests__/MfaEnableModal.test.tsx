import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { MfaGenerate } from "@/client";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import MfaEnableModal from "../MfaEnableModal";

vi.mock("qrcode");

const mockMfaData: MfaGenerate = {
  link: "otpauth://totp/ShellHub:user@example.com?secret=ABCD1234&issuer=ShellHub",
  secret: "ABCD1234",
  recovery_codes: ["code1", "code2", "code3", "code4", "code5", "code6"],
};

describe("MfaEnableModal", () => {
  const onClose = vi.fn();
  const onSuccess = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    server.use(
      http.get("*/api/user/mfa/generate", () => HttpResponse.json(mockMfaData)),
      http.put(
        "*/api/user/mfa/enable",
        () => new HttpResponse(null, { status: 204 }),
      ),
      http.patch("*/api/users", () => new HttpResponse(null, { status: 204 })),
    );
  });

  describe("Step 1: Recovery Email", () => {
    it("shows error when email is already in use (409)", async () => {
      const user = userEvent.setup();
      server.use(
        http.patch("*/api/users", () => HttpResponse.json({}, { status: 409 })),
      );

      render(
        <MfaEnableModal
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail={null}
        />,
      );

      const emailInput = screen.getByPlaceholderText(/recovery@example\.com/i);
      await user.type(emailInput, "duplicate@example.com");

      const nextButton = screen.getByRole("button", { name: /save/i });
      await user.click(nextButton);

      await waitFor(() => {
        expect(screen.getByText(/email already in use/i)).toBeInTheDocument();
      });
    });
  });

  describe("Step 2: Recovery Codes", () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      render(
        <MfaEnableModal
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />,
      );

      const continueButton = screen.getByRole("button", { name: /continue/i });
      await user.click(continueButton);
      await waitFor(() => {
        expect(screen.getByText(/Save Recovery Codes/i)).toBeInTheDocument();
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
  });

  describe("Step 3: QR Code and Verification", () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      render(
        <MfaEnableModal
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />,
      );

      const continueButton = screen.getByRole("button", { name: /continue/i });
      await user.click(continueButton);
      await waitFor(() => {
        expect(screen.getByText(/Save Recovery Codes/i)).toBeInTheDocument();
      });

      const checkbox = screen.getByRole("checkbox");
      await user.click(checkbox);
      const nextButton = screen.getByRole("button", { name: /next/i });
      await user.click(nextButton);
    });

    it("displays the QR code and the secret for manual entry", async () => {
      await waitFor(() => {
        expect(screen.getByText(/scan this qr code/i)).toBeInTheDocument();
      });
      expect(
        screen.getByRole("img", { name: "MFA QR code" }),
      ).toBeInTheDocument();
      expect(screen.getByDisplayValue(mockMfaData.secret)).toBeInTheDocument();
    });

    it("shows error on invalid OTP", async () => {
      const user = userEvent.setup();
      server.use(
        http.put("*/api/user/mfa/enable", () =>
          HttpResponse.json({}, { status: 403 }),
        ),
      );

      await waitFor(() => {
        expect(screen.getByText(/scan this qr code/i)).toBeInTheDocument();
      });

      const inputs = screen.getAllByRole("textbox");
      const otpInputs = inputs.filter(
        (input) => input.getAttribute("maxLength") === "1",
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
        expect(
          screen.getByText(/invalid verification code/i),
        ).toBeInTheDocument();
      });
    });
  });

  it("rewinds to step 1 when the modal is cancelled and reopened", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <MfaEnableModal
        open={true}
        onClose={onClose}
        onSuccess={onSuccess}
        currentRecoveryEmail="recovery@example.com"
      />,
    );

    await user.click(screen.getByRole("button", { name: /continue/i }));
    await waitFor(() => {
      expect(screen.getByText(/Save Recovery Codes/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /cancel/i }));
    expect(onClose).toHaveBeenCalled();

    rerender(
      <MfaEnableModal
        open={false}
        onClose={onClose}
        onSuccess={onSuccess}
        currentRecoveryEmail="recovery@example.com"
      />,
    );
    rerender(
      <MfaEnableModal
        open={true}
        onClose={onClose}
        onSuccess={onSuccess}
        currentRecoveryEmail="recovery@example.com"
      />,
    );

    await waitFor(() => {
      expect(screen.getByText(/Confirm Recovery Email/i)).toBeInTheDocument();
    });
    expect(screen.queryByText(/Save Recovery Codes/i)).not.toBeInTheDocument();
  });

  describe("Error Handling", () => {
    it("handles API errors when generating MFA codes", async () => {
      const user = userEvent.setup();
      server.use(
        http.get("*/api/user/mfa/generate", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );

      render(
        <MfaEnableModal
          open={true}
          onClose={onClose}
          onSuccess={onSuccess}
          currentRecoveryEmail="recovery@example.com"
        />,
      );

      const continueButton = screen.getByRole("button", { name: /continue/i });
      await user.click(continueButton);

      await waitFor(() => {
        expect(screen.getByText(/failed to generate/i)).toBeInTheDocument();
      });
    });
  });
});
