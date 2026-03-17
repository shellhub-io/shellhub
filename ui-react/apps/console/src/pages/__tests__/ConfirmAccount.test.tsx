import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { useSignUpStore } from "../../stores/signUpStore";
import ConfirmAccount from "../ConfirmAccount";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual };
});

vi.mock("../../client", () => ({
  registerUser: vi.fn(),
  resendEmail: vi.fn(),
  getValidateAccount: vi.fn(),
}));

import { resendEmail as apiResendEmail } from "../../client";
const mockedResendEmail = vi.mocked(apiResendEmail);

type SdkResponse<T = unknown> = { data: T; request: Request; response: Response };

function mockSdkResponse<T>(data: T): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(),
  };
}

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function renderConfirmAccount(username?: string) {
  const search = username !== undefined ? `?username=${encodeURIComponent(username)}` : "";
  return render(
    <MemoryRouter initialEntries={[`/confirm-account${search}`]}>
      <ConfirmAccount />
    </MemoryRouter>,
  );
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  mockedResendEmail.mockReset();
  useSignUpStore.setState({ resendLoading: false, resendError: null });
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("ConfirmAccount", () => {
  describe("rendering", () => {
    it("renders the heading and resend button", () => {
      renderConfirmAccount("admin");
      expect(screen.getByText(/account activation required/i)).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /resend email/i })).toBeInTheDocument();
    });

    it("renders a back-to-login link", () => {
      renderConfirmAccount("admin");
      expect(screen.getByRole("link", { name: /login/i })).toBeInTheDocument();
    });

    it("redirects to /login when no username is provided", () => {
      renderConfirmAccount();
      expect(screen.queryByText(/account activation required/i)).not.toBeInTheDocument();
    });

    it("enables the button when a username is provided", () => {
      renderConfirmAccount("admin");
      expect(screen.getByRole("button", { name: /resend email/i })).not.toBeDisabled();
    });
  });

  describe("resend email", () => {
    it("calls resendEmail with the username and shows success message", async () => {
      mockedResendEmail.mockResolvedValue(mockSdkResponse(undefined));

      renderConfirmAccount("admin");
      await userEvent.click(screen.getByRole("button", { name: /resend email/i }));

      expect(mockedResendEmail).toHaveBeenCalledWith({ body: { username: "admin" }, throwOnError: true });
      await waitFor(() =>
        expect(screen.getByText(/confirmation email sent successfully/i)).toBeInTheDocument(),
      );
    });

    it("shows an error message on failure", async () => {
      mockedResendEmail.mockRejectedValue(new Error("500"));

      renderConfirmAccount("admin");
      await userEvent.click(screen.getByRole("button", { name: /resend email/i }));

      await waitFor(() =>
        expect(screen.getByText(/failed to resend email/i)).toBeInTheDocument(),
      );
    });

    it("shows Sending... and disables the button while the request is in flight", async () => {
      let resolveResend!: () => void;
      mockedResendEmail.mockReturnValue(
        new Promise<SdkResponse>((resolve) => {
          resolveResend = () => resolve(mockSdkResponse(undefined));
        }),
      );

      renderConfirmAccount("admin");
      const clickPromise = userEvent.click(
        screen.getByRole("button", { name: /resend email/i }),
      );

      await waitFor(() =>
        expect(screen.getByText(/sending/i)).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: /sending/i })).toBeDisabled();

      resolveResend();
      await clickPromise;
    });
  });
});
