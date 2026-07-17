import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import MfaDisableDialog from "../MfaDisableDialog";

vi.mock("@/client", () => ({
  disableMfa: vi.fn(),
  requestResetMfa: vi.fn(),
}));

vi.mock("@/hooks/useFocusTrap", () => ({ useFocusTrap: vi.fn() }));

import { disableMfa, requestResetMfa } from "@/client";
import { useAuthStore } from "@/stores/authStore";

const mockedDisableMfa = vi.mocked(disableMfa);
const mockedRequestResetMfa = vi.mocked(requestResetMfa);

type SdkResponse<T = unknown> = {
  data: T;
  request: Request;
  response: Response;
};

function mockSdkResponse<T>(data: T): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(),
  };
}

describe("MfaDisableDialog", () => {
  const onClose = vi.fn();
  const onSuccess = vi.fn();

  function renderDialog(open = true) {
    const user = userEvent.setup();
    render(
      <MfaDisableDialog open={open} onClose={onClose} onSuccess={onSuccess} />,
    );
    return user;
  }

  async function fillTotpCode(
    user: ReturnType<typeof userEvent.setup>,
    code = "123456",
  ) {
    const inputs = screen.getAllByRole("textbox");
    const otpInputs = inputs.filter(
      (input) => input.getAttribute("maxLength") === "1",
    );
    for (let i = 0; i < code.length; i++) {
      await user.type(otpInputs[i], code[i]);
    }
  }

  beforeEach(() => {
    vi.clearAllMocks();
    mockedDisableMfa.mockResolvedValue(mockSdkResponse(undefined));
  });

  describe("Mode Switching", () => {
    it("defaults to TOTP mode", () => {
      renderDialog();

      expect(screen.getByText(/Verification Code/i)).toBeInTheDocument();
      expect(
        screen.getByText(/Use recovery code instead/i),
      ).toBeInTheDocument();
    });

    it("switches to recovery code mode", async () => {
      const user = renderDialog();

      await user.click(screen.getByText(/use recovery code/i));

      expect(screen.getByPlaceholderText(/recovery code/i)).toBeInTheDocument();
    });

    it("switches back to TOTP mode from recovery", async () => {
      const user = renderDialog();

      await user.click(screen.getByText(/use recovery code/i));
      await user.click(screen.getByText(/use authenticator/i));

      expect(screen.getByText(/Verification Code/i)).toBeInTheDocument();
    });
  });

  describe("TOTP Mode Validation", () => {
    it("requires all 6 digits before enabling submit", async () => {
      const user = renderDialog();

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      expect(disableButton).toBeDisabled();

      await fillTotpCode(user);

      expect(disableButton).toBeEnabled();
    });

    it("successfully disables MFA with valid TOTP", async () => {
      const user = renderDialog();

      await fillTotpCode(user);

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => {
        expect(mockedDisableMfa).toHaveBeenCalledWith({
          body: { code: "123456" },
          throwOnError: true,
        });
        expect(onSuccess).toHaveBeenCalled();
        expect(onClose).toHaveBeenCalled();
      });
    });

    it("shows error on invalid TOTP", async () => {
      mockedDisableMfa.mockRejectedValue(new Error("Invalid code"));
      const user = renderDialog();

      await fillTotpCode(user, "999999");

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/Invalid verification code/i),
        ).toBeInTheDocument();
      });
      expect(onSuccess).not.toHaveBeenCalled();
    });
  });

  describe("Recovery Code Mode Validation", () => {
    it("requires recovery code before enabling submit", async () => {
      const user = renderDialog();

      await user.click(screen.getByText(/use recovery code/i));

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      expect(disableButton).toBeDisabled();

      await user.type(
        screen.getByPlaceholderText(/recovery code/i),
        "abc123xyz",
      );

      expect(disableButton).toBeEnabled();
    });

    it("successfully disables MFA with valid recovery code", async () => {
      const user = renderDialog();

      await user.click(screen.getByText(/use recovery code/i));
      await user.type(
        screen.getByPlaceholderText(/recovery code/i),
        "valid-recovery-code",
      );

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => {
        expect(mockedDisableMfa).toHaveBeenCalledWith({
          body: { recovery_code: "valid-recovery-code" },
          throwOnError: true,
        });
        expect(onSuccess).toHaveBeenCalled();
        expect(onClose).toHaveBeenCalled();
      });
    });

    it("shows error on invalid recovery code", async () => {
      mockedDisableMfa.mockRejectedValue(new Error("Invalid recovery code"));
      const user = renderDialog();

      await user.click(screen.getByText(/use recovery code/i));
      await user.type(
        screen.getByPlaceholderText(/recovery code/i),
        "invalid-code",
      );

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => {
        expect(screen.getByText(/Invalid recovery code/i)).toBeInTheDocument();
      });
      expect(onSuccess).not.toHaveBeenCalled();
    });
  });

  describe("Dialog Behavior", () => {
    it("closes when cancel button is clicked", async () => {
      const user = renderDialog();

      await user.click(screen.getByText(/cancel/i));

      expect(onClose).toHaveBeenCalled();
      expect(onSuccess).not.toHaveBeenCalled();
    });

    it("closes when clicking outside (backdrop)", () => {
      renderDialog();

      const dialog = document.querySelector("dialog") as HTMLElement;
      fireEvent.mouseDown(dialog);
      fireEvent.click(dialog);
      expect(onClose).toHaveBeenCalled();
    });

    it("does not render when open is false", () => {
      const { container } = render(
        <MfaDisableDialog
          open={false}
          onClose={onClose}
          onSuccess={onSuccess}
        />,
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe("Loading State", () => {
    it("disables submit button while submitting", async () => {
      let resolveDisable: (v: SdkResponse) => void;
      mockedDisableMfa.mockReturnValue(
        new Promise<SdkResponse>((resolve) => {
          resolveDisable = resolve;
        }),
      );

      const user = renderDialog();

      await fillTotpCode(user);

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      await user.click(disableButton);

      expect(disableButton).toBeDisabled();

      resolveDisable!(mockSdkResponse(undefined));

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalled();
      });
    });
  });

  describe("Error Handling", () => {
    it("shows error after failed TOTP submit and switches to recovery mode", async () => {
      mockedDisableMfa.mockRejectedValue(new Error("Invalid code"));
      const user = renderDialog();

      await fillTotpCode(user, "999999");

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/Invalid verification code/i),
        ).toBeInTheDocument();
      });

      await user.click(screen.getByText(/use recovery code instead/i));

      expect(screen.getByPlaceholderText(/recovery code/i)).toBeInTheDocument();
    });
  });

  describe("Email-Reset Mode", () => {
    beforeEach(() => {
      useAuthStore.setState({ user: "admin" });
      mockedRequestResetMfa.mockResolvedValue(
        mockSdkResponse({ token: "reset-token" }),
      );
    });

    async function navigateToEmailReset(
      user: ReturnType<typeof userEvent.setup>,
    ) {
      await user.click(screen.getByText(/use recovery code/i));
      await user.click(screen.getByText(/request email reset/i));
    }

    async function requestCodes(user: ReturnType<typeof userEvent.setup>) {
      await user.click(
        screen.getByRole("button", { name: /send verification codes/i }),
      );
      await waitFor(() => {
        expect(screen.getByText(/Emails Sent!/i)).toBeInTheDocument();
      });
    }

    async function fillEmailOtpInputs(
      user: ReturnType<typeof userEvent.setup>,
    ) {
      const mainInputs = screen.getAllByLabelText(/main email code character/i);
      for (let i = 0; i < mainInputs.length; i++) {
        await user.type(mainInputs[i], String.fromCharCode(65 + i));
      }

      const recoveryInputs = screen.getAllByLabelText(
        /recovery email code character/i,
      );
      for (let i = 0; i < recoveryInputs.length; i++) {
        await user.type(recoveryInputs[i], String(i + 1));
      }
    }

    it("shows OTP inputs after requesting codes", async () => {
      const user = renderDialog();

      await navigateToEmailReset(user);
      await requestCodes(user);

      expect(
        screen.getAllByLabelText(/main email code character/i),
      ).toHaveLength(5);
      expect(
        screen.getAllByLabelText(/recovery email code character/i),
      ).toHaveLength(5);
      expect(
        screen.getByRole("button", { name: /disable mfa/i }),
      ).toBeInTheDocument();
    });

    it("submits email codes and calls onSuccess/onClose", async () => {
      const user = renderDialog();

      await navigateToEmailReset(user);
      await requestCodes(user);
      await fillEmailOtpInputs(user);

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => {
        expect(mockedDisableMfa).toHaveBeenCalledWith({
          body: { main_email_code: "ABCDE", recovery_email_code: "12345" },
          throwOnError: true,
        });
        expect(onSuccess).toHaveBeenCalled();
        expect(onClose).toHaveBeenCalled();
      });
    });

    it("shows error and resets OTP inputs on failure", async () => {
      mockedDisableMfa.mockRejectedValue(new Error("Invalid codes"));
      const user = renderDialog();

      await navigateToEmailReset(user);
      await requestCodes(user);
      await fillEmailOtpInputs(user);

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/Invalid email verification codes/i),
        ).toBeInTheDocument();
      });
      expect(onSuccess).not.toHaveBeenCalled();
      expect(onClose).not.toHaveBeenCalled();

      const mainInputs = screen.getAllByLabelText(/main email code character/i);
      const recoveryInputs = screen.getAllByLabelText(
        /recovery email code character/i,
      );
      for (const input of [...mainInputs, ...recoveryInputs]) {
        expect(input).toHaveValue("");
      }
    });

    it("resets email-reset state when switching back to recovery", async () => {
      const user = renderDialog();

      await navigateToEmailReset(user);

      expect(
        screen.getByRole("button", { name: /send verification codes/i }),
      ).toBeInTheDocument();

      await user.click(screen.getByText(/use recovery code/i));

      expect(screen.getByPlaceholderText(/recovery code/i)).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /disable mfa/i }),
      ).toBeInTheDocument();
    });
  });
});
