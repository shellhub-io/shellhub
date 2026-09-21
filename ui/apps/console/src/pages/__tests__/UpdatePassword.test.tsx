import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import UpdatePassword from "../UpdatePassword";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

function renderWithParams(search = "?id=uid123&token=tok456") {
  return render(
    <MemoryRouter initialEntries={[`/update-password${search}`]}>
      <Routes>
        <Route path="/update-password" element={<UpdatePassword />} />
        <Route path="/login" element={<div>Login Page</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  server.use(
    http.post(
      "*/api/user/:uid/update_password",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("UpdatePassword", () => {
  describe("uid/token guard", () => {
    it.each(["?id=&token=", "?token=tok456", "?id=uid123"])(
      "renders an error card with a link instead of the form for %s",
      (search) => {
        renderWithParams(search);

        expect(screen.getByText(/invalid reset link/i)).toBeInTheDocument();
        expect(
          screen.getByRole("link", { name: /request a new reset link/i }),
        ).toBeInTheDocument();
        expect(
          screen.queryByRole("button", { name: /update password/i }),
        ).not.toBeInTheDocument();
      },
    );
  });

  describe("validation — errors appear after blur", () => {
    it("shows a too-short error only after the password field is blurred", async () => {
      const user = userEvent.setup();
      renderWithParams();

      expect(screen.queryByText(/password must be/i)).not.toBeInTheDocument();

      await user.type(screen.getByLabelText(/^new password$/i), "abc");
      await user.tab();

      expect(await screen.findByText(/password must be/i)).toBeInTheDocument();
    });

    it("shows a mismatch error only after confirmPassword is blurred", async () => {
      const user = userEvent.setup();
      renderWithParams();

      expect(
        screen.queryByText(/passwords do not match/i),
      ).not.toBeInTheDocument();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(
        screen.getByLabelText(/^confirm password$/i),
        "Different1",
      );
      await user.tab();

      expect(
        await screen.findByText(/passwords do not match/i),
      ).toBeInTheDocument();
    });
  });

  describe("submit button gate", () => {
    it("enables the submit button only when both password fields contain valid, matching values", async () => {
      const user = userEvent.setup();
      renderWithParams();

      expect(
        screen.getByRole("button", { name: /update password/i }),
      ).toBeDisabled();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(
        screen.getByLabelText(/^confirm password$/i),
        "Secret123",
      );

      expect(
        screen.getByRole("button", { name: /update password/i }),
      ).toBeEnabled();
    });

    it("keeps the submit button disabled when passwords match but are too short", async () => {
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "abc");
      await user.type(screen.getByLabelText(/^confirm password$/i), "abc");

      expect(
        screen.getByRole("button", { name: /update password/i }),
      ).toBeDisabled();
    });
  });

  describe("successful submission", () => {
    it("navigates to /login with a success notice after successful submission", async () => {
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(
        screen.getByLabelText(/^confirm password$/i),
        "Secret123",
      );
      await user.click(
        screen.getByRole("button", { name: /update password/i }),
      );

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          "/login",
          expect.objectContaining({
            state: expect.objectContaining({ notice: expect.any(String) }),
          }),
        ),
      );
    });
  });

  describe("API failure", () => {
    it("shows a generic error message when the API call fails", async () => {
      server.use(
        http.post("*/api/user/:uid/update_password", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(
        screen.getByLabelText(/^confirm password$/i),
        "Secret123",
      );
      await user.click(
        screen.getByRole("button", { name: /update password/i }),
      );

      expect(await screen.findByRole("alert")).toBeInTheDocument();
    });
  });
});
