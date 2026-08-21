import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { useSignUpStore } from "@/stores/signUpStore";
import ConfirmAccount from "../ConfirmAccount";
import { mockSdkResponse, type SdkResponse } from "@/tests/sdk";

const mockResendEmail = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, resendEmail: mockResendEmail };
});

function renderConfirmAccount(username?: string) {
  const search =
    username !== undefined ? `?username=${encodeURIComponent(username)}` : "";
  return render(
    <MemoryRouter initialEntries={[`/confirm-account${search}`]}>
      <ConfirmAccount />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  mockResendEmail.mockReset();
  useSignUpStore.setState({ resendLoading: false, resendError: null });
});

describe("ConfirmAccount", () => {
  describe("rendering", () => {
    it("renders the heading and resend button", () => {
      renderConfirmAccount("admin");
      expect(
        screen.getByText(/account activation required/i),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /resend email/i }),
      ).toBeInTheDocument();
    });

    it("renders a back-to-login link", () => {
      renderConfirmAccount("admin");
      expect(screen.getByRole("link", { name: /login/i })).toBeInTheDocument();
    });

    it("redirects to /login when no username is provided", () => {
      renderConfirmAccount();
      expect(
        screen.queryByText(/account activation required/i),
      ).not.toBeInTheDocument();
    });

    it("enables the button when a username is provided", () => {
      renderConfirmAccount("admin");
      expect(
        screen.getByRole("button", { name: /resend email/i }),
      ).not.toBeDisabled();
    });
  });

  describe("resend email", () => {
    it("calls resendEmail with the username and shows success message", async () => {
      mockResendEmail.mockResolvedValue(mockSdkResponse(undefined));

      renderConfirmAccount("admin");
      await userEvent.click(
        screen.getByRole("button", { name: /resend email/i }),
      );

      expect(mockResendEmail).toHaveBeenCalledWith({
        body: { username: "admin" },
        throwOnError: true,
      });
      await waitFor(() =>
        expect(
          screen.getByText(/confirmation email sent successfully/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows an error message on failure", async () => {
      mockResendEmail.mockRejectedValue(new Error("500"));

      renderConfirmAccount("admin");
      await userEvent.click(
        screen.getByRole("button", { name: /resend email/i }),
      );

      await waitFor(() =>
        expect(screen.getByText(/failed to resend email/i)).toBeInTheDocument(),
      );
    });

    it("shows Sending... and disables the button while the request is in flight", async () => {
      let resolveResend!: () => void;
      mockResendEmail.mockReturnValue(
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
