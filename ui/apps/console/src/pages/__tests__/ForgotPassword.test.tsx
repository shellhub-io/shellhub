import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import ForgotPassword from "../ForgotPassword";

function renderForgotPassword() {
  return render(
    <MemoryRouter>
      <ForgotPassword />
    </MemoryRouter>,
  );
}
beforeEach(() => {
  vi.clearAllMocks();
  server.use(
    http.post(
      "*/api/user/recover_password",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("ForgotPassword", () => {
  describe("initial state", () => {
    it("does not show an account error before the field is touched", () => {
      renderForgotPassword();
      expect(
        screen.queryByText(/enter a valid username or email/i),
      ).not.toBeInTheDocument();
      expect(screen.queryByText(/is required/i)).not.toBeInTheDocument();
    });

    it("disables the submit button when the form is empty", () => {
      renderForgotPassword();
      expect(
        screen.getByRole("button", { name: /reset password/i }),
      ).toBeDisabled();
    });
  });

  describe("field validation", () => {
    it("shows an error after blurring with an invalid account value", async () => {
      const user = userEvent.setup();
      renderForgotPassword();

      const input = screen.getByLabelText(/username or email/i);
      await user.type(input, "!!");
      await user.tab();

      expect(
        await screen.findByText(/enter a valid username or email/i),
      ).toBeInTheDocument();
    });

    it("keeps the submit button disabled when the account is invalid", async () => {
      const user = userEvent.setup();
      renderForgotPassword();

      await user.type(screen.getByLabelText(/username or email/i), "!!");
      await user.tab();

      await screen.findByText(/enter a valid username or email/i);
      expect(
        screen.getByRole("button", { name: /reset password/i }),
      ).toBeDisabled();
    });
  });

  describe("valid submission", () => {
    it("enables the submit button once a valid account is entered", async () => {
      const user = userEvent.setup();
      renderForgotPassword();

      await user.type(screen.getByLabelText(/username or email/i), "alice");

      expect(
        screen.getByRole("button", { name: /reset password/i }),
      ).toBeEnabled();
    });

    it("shows the sent view after a successful submission", async () => {
      const user = userEvent.setup();
      renderForgotPassword();

      await user.type(screen.getByLabelText(/username or email/i), "alice");
      await user.click(screen.getByRole("button", { name: /reset password/i }));

      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(screen.getByText(/check your inbox/i)).toBeInTheDocument();
    });

    it("shows the sent view even when the API call fails (anti-enumeration)", async () => {
      server.use(
        http.post("*/api/user/recover_password", () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      );
      const user = userEvent.setup();
      renderForgotPassword();

      await user.type(screen.getByLabelText(/username or email/i), "alice");
      await user.click(screen.getByRole("button", { name: /reset password/i }));

      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(screen.getByText(/check your inbox/i)).toBeInTheDocument();
    });
  });
});
