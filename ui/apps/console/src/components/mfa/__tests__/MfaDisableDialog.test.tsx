import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Route } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockInfo } from "@/tests/factories";
import { renderWithSignInPage } from "@/tests/renderWithSignInPage";
import { createTestWrapper } from "@/tests/wrapper";
import MfaDisableDialog from "../MfaDisableDialog";

describe("MfaDisableDialog", () => {
  const onClose = vi.fn();
  const onSuccess = vi.fn();

  function renderDialog(open = true) {
    const user = userEvent.setup();
    renderWithSignInPage(
      <Route
        path="/"
        element={
          <MfaDisableDialog
            open={open}
            onClose={onClose}
            onSuccess={onSuccess}
          />
        }
      />,
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

  async function navigateToEmailReset(
    user: ReturnType<typeof userEvent.setup>,
  ) {
    await user.click(
      screen.getByRole("button", { name: /use recovery code/i }),
    );
    await user.click(
      screen.getByRole("button", { name: /request email reset/i }),
    );
  }

  async function requestCodes(user: ReturnType<typeof userEvent.setup>) {
    await user.click(
      screen.getByRole("button", { name: /send verification codes/i }),
    );
    await waitFor(() => {
      expect(screen.getByText(/Emails Sent!/i)).toBeInTheDocument();
    });
  }

  async function backToAuthenticatorCode(
    user: ReturnType<typeof userEvent.setup>,
  ) {
    await user.click(
      screen.getByRole("button", { name: /use authenticator code/i }),
    );
  }

  beforeEach(() => {
    vi.clearAllMocks();
    seedAuthStore();
    server.use(
      http.put(
        "*/api/user/mfa/disable",
        () => new HttpResponse(null, { status: 204 }),
      ),
      http.post("*/api/user/mfa/reset", () =>
        HttpResponse.json({ token: "reset-token" }),
      ),
    );
  });

  describe("Mode Switching", () => {
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

  describe("reopening", () => {
    function reopenable() {
      const user = userEvent.setup();
      const tree = (open: boolean) => (
        <MfaDisableDialog open={open} onClose={onClose} onSuccess={onSuccess} />
      );
      const { rerender } = render(tree(true), {
        wrapper: createTestWrapper({ initialEntries: ["/"] }),
      });
      const closeAndReopen = () => {
        rerender(tree(false));
        rerender(tree(true));
      };
      return { user, closeAndReopen };
    }

    it("clears the last error", async () => {
      server.use(
        http.put("*/api/user/mfa/disable", () =>
          HttpResponse.json({}, { status: 403 }),
        ),
      );
      const { user, closeAndReopen } = reopenable();
      await fillTotpCode(user, "999999");
      await user.click(screen.getByRole("button", { name: /disable mfa/i }));
      expect(
        await screen.findByText(/Invalid verification code/i),
      ).toBeInTheDocument();

      closeAndReopen();

      expect(
        screen.queryByText(/Invalid verification code/i),
      ).not.toBeInTheDocument();
    });

    it("goes back to the authenticator code", async () => {
      const { user, closeAndReopen } = reopenable();
      await user.click(screen.getByText(/use recovery code/i));

      closeAndReopen();

      expect(
        screen.queryByPlaceholderText(/recovery code/i),
      ).not.toBeInTheDocument();
      expect(screen.getByText(/Verification Code/i)).toBeInTheDocument();
    });

    it("stays on the email codes once they were sent", async () => {
      const { user, closeAndReopen } = reopenable();
      await navigateToEmailReset(user);
      await requestCodes(user);

      closeAndReopen();

      expect(screen.getByText(/Emails Sent!/i)).toBeInTheDocument();
      expect(
        screen.getByLabelText(/main email code character 1/i),
      ).toBeInTheDocument();
    });

    it("asks for the authenticator code again after MFA was disabled with it", async () => {
      const { user, closeAndReopen } = reopenable();
      await navigateToEmailReset(user);
      await requestCodes(user);
      await backToAuthenticatorCode(user);
      await fillTotpCode(user);
      await user.click(screen.getByRole("button", { name: /disable mfa/i }));
      await waitFor(() => expect(onSuccess).toHaveBeenCalled());

      closeAndReopen();

      expect(screen.queryByText(/Emails Sent!/i)).not.toBeInTheDocument();
      expect(screen.getByText(/Verification Code/i)).toBeInTheDocument();
    });

    it("comes back to the email codes after leaving them for the authenticator code", async () => {
      const { user, closeAndReopen } = reopenable();
      await navigateToEmailReset(user);
      await requestCodes(user);
      await backToAuthenticatorCode(user);

      closeAndReopen();

      expect(screen.getByText(/Emails Sent!/i)).toBeInTheDocument();
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
        expect(onSuccess).toHaveBeenCalled();
        expect(onClose).toHaveBeenCalled();
      });
    });

    it("shows error on invalid TOTP", async () => {
      server.use(
        http.put("*/api/user/mfa/disable", () =>
          HttpResponse.json({}, { status: 403 }),
        ),
      );
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

    it("names the lockout a wrong TOTP starts", async () => {
      server.use(
        http.put("*/api/user/mfa/disable", () =>
          HttpResponse.json(
            { message: "too many wrong attempts" },
            {
              status: 429,
              headers: { "X-Account-Lockout": String(Math.floor(Date.now() / 1000) + 270) },
            },
          ),
        ),
      );
      const user = renderDialog();

      await fillTotpCode(user, "999999");

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      expect(
        await screen.findByText("Too many wrong attempts. Try again in 4 minutes."),
      ).toBeInTheDocument();
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
        expect(onSuccess).toHaveBeenCalled();
        expect(onClose).toHaveBeenCalled();
      });
    });

    it("shows error on invalid recovery code", async () => {
      server.use(
        http.put("*/api/user/mfa/disable", () =>
          HttpResponse.json({}, { status: 403 }),
        ),
      );
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

  describe("Loading State", () => {
    it("disables submit button while submitting", async () => {
      server.use(
        http.put("*/api/user/mfa/disable", () => new Promise(() => {})),
      );

      const user = renderDialog();

      await fillTotpCode(user);

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      await user.click(disableButton);

      await waitFor(() => {
        expect(disableButton).toBeDisabled();
      });
    });
  });

  describe("Email-Reset Mode", () => {
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

    it("says a reset was already asked for when the request is throttled", async () => {
      server.use(
        http.post("*/api/user/mfa/reset", () =>
          HttpResponse.json({ message: "too many requests" }, { status: 429 }),
        ),
      );
      const user = renderDialog();

      await navigateToEmailReset(user);
      await user.click(
        screen.getByRole("button", { name: /send verification codes/i }),
      );

      expect(
        await screen.findByText(/already requested in the last minute/i),
      ).toBeInTheDocument();
      expect(screen.queryByText(/Emails Sent!/i)).not.toBeInTheDocument();
    });

    it("submits email codes, then signs out on the sign-in page", async () => {
      server.use(http.get("*/info", () => HttpResponse.json(mockInfo())));
      const user = renderDialog();

      await navigateToEmailReset(user);
      await requestCodes(user);
      await fillEmailOtpInputs(user);

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      expect(
        await screen.findByText("MFA disabled. Sign in again."),
      ).toBeInTheDocument();
      expect(useAuthStore.getState().token).toBeNull();
      expect(onSuccess).not.toHaveBeenCalled();
      expect(onClose).not.toHaveBeenCalled();
    });

    it("shows error and resets OTP inputs on failure", async () => {
      server.use(
        http.put("*/api/user/mfa/disable", () =>
          HttpResponse.json({}, { status: 403 }),
        ),
      );
      const user = renderDialog();

      await navigateToEmailReset(user);
      await requestCodes(user);
      await fillEmailOtpInputs(user);

      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/Three wrong attempts cancel the codes/i),
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

    it("disables MFA with the authenticator code after the codes were sent", async () => {
      let body: unknown;
      server.use(
        http.put("*/api/user/mfa/disable", async ({ request }) => {
          body = await request.json();
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const user = renderDialog();

      await navigateToEmailReset(user);
      await requestCodes(user);
      await backToAuthenticatorCode(user);
      await fillTotpCode(user);
      await user.click(screen.getByRole("button", { name: /disable mfa/i }));

      await waitFor(() => expect(onSuccess).toHaveBeenCalled());
      expect(body).toEqual({ code: "123456" });
    });

    it("drops the rejected email codes error when going back to the authenticator code", async () => {
      server.use(
        http.put("*/api/user/mfa/disable", () =>
          HttpResponse.json({}, { status: 403 }),
        ),
      );
      const user = renderDialog();

      await navigateToEmailReset(user);
      await requestCodes(user);
      await fillEmailOtpInputs(user);
      await user.click(screen.getByRole("button", { name: /disable mfa/i }));
      expect(
        await screen.findByText(/Three wrong attempts cancel the codes/i),
      ).toBeInTheDocument();
      await backToAuthenticatorCode(user);

      expect(
        screen.queryByText(/Three wrong attempts cancel the codes/i),
      ).not.toBeInTheDocument();
    });

    it("keeps the sent codes when coming back to the email reset", async () => {
      const user = renderDialog();

      await navigateToEmailReset(user);
      await requestCodes(user);
      await backToAuthenticatorCode(user);
      await navigateToEmailReset(user);

      expect(screen.getByText(/Emails Sent!/i)).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /send verification codes/i }),
      ).not.toBeInTheDocument();
    });

    it("returns to the recovery code before any codes are sent", async () => {
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
