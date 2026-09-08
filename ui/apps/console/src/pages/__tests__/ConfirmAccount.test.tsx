import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useSignUpStore } from "@/stores/signUpStore";
import ConfirmAccount from "../ConfirmAccount";

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
  vi.clearAllMocks();
  useSignUpStore.setState({ resendLoading: false, resendError: null });
  server.use(
    http.post(
      "*/api/user/resend_email",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
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
    it("shows success message after resending", async () => {
      renderConfirmAccount("admin");
      await userEvent.click(
        screen.getByRole("button", { name: /resend email/i }),
      );

      await waitFor(() =>
        expect(
          screen.getByText(/confirmation email sent successfully/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows an error message on failure", async () => {
      server.use(
        http.post("*/api/user/resend_email", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );

      renderConfirmAccount("admin");
      await userEvent.click(
        screen.getByRole("button", { name: /resend email/i }),
      );

      await waitFor(() =>
        expect(screen.getByText(/failed to resend email/i)).toBeInTheDocument(),
      );
    });

    it("shows Sending... and disables the button while the request is in flight", async () => {
      let resolveHandler!: () => void;
      server.use(
        http.post(
          "*/api/user/resend_email",
          () =>
            new Promise<Response>((resolve) => {
              resolveHandler = () =>
                resolve(new HttpResponse(null, { status: 204 }));
            }),
        ),
      );

      renderConfirmAccount("admin");
      const clickPromise = userEvent.click(
        screen.getByRole("button", { name: /resend email/i }),
      );

      await waitFor(() =>
        expect(screen.getByText(/sending/i)).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: /sending/i })).toBeDisabled();

      resolveHandler();
      await clickPromise;
    });
  });
});
