import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import UpdatePassword from "../UpdatePassword";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/client", () => ({
  updateRecoverPassword: vi.fn(),
}));

import { updateRecoverPassword } from "@/client";
const mockedUpdate = vi.mocked(updateRecoverPassword);

function mockSdkResponse<T>(data: T) {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(),
  };
}

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
  mockNavigate.mockReset();
  mockedUpdate.mockReset();
  mockedUpdate.mockResolvedValue(mockSdkResponse(undefined));
});

describe("UpdatePassword", () => {
  describe("uid/token guard", () => {
    it("renders an error card with a link when uid and token are missing", () => {
      renderWithParams("?id=&token=");

      expect(screen.getByText(/invalid reset link/i)).toBeInTheDocument();
      expect(
        screen.getByRole("link", { name: /request a new reset link/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /update password/i }),
      ).not.toBeInTheDocument();
    });

    it("renders an error card when only uid is missing", () => {
      renderWithParams("?token=tok456");

      expect(screen.getByText(/invalid reset link/i)).toBeInTheDocument();
    });

    it("renders an error card when only token is missing", () => {
      renderWithParams("?id=uid123");

      expect(screen.getByText(/invalid reset link/i)).toBeInTheDocument();
    });
  });

  describe("validation — no errors before blur", () => {
    it("shows no password error on initial render", () => {
      renderWithParams();

      expect(
        screen.queryByText(/password must be/i),
      ).not.toBeInTheDocument();
    });

    it("shows no mismatch error on initial render", () => {
      renderWithParams();

      expect(
        screen.queryByText(/passwords do not match/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("validation — errors appear after blur", () => {
    it("shows a too-short error after the password field is blurred with a short value", async () => {
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "abc");
      await user.tab();

      expect(
        await screen.findByText(/password must be/i),
      ).toBeInTheDocument();
    });

    it("shows a mismatch error after confirmPassword is blurred with a non-matching value", async () => {
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(screen.getByLabelText(/^confirm password$/i), "Different1");
      await user.tab();

      expect(
        await screen.findByText(/passwords do not match/i),
      ).toBeInTheDocument();
    });
  });

  describe("submit button gate", () => {
    it("disables the submit button on initial render", () => {
      renderWithParams();

      expect(
        screen.getByRole("button", { name: /update password/i }),
      ).toBeDisabled();
    });

    it("enables the submit button only when both password fields contain valid, matching values", async () => {
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(screen.getByLabelText(/^confirm password$/i), "Secret123");

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
    it("calls updateRecoverPassword with uid, token, and password on valid submit", async () => {
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(screen.getByLabelText(/^confirm password$/i), "Secret123");
      await user.click(screen.getByRole("button", { name: /update password/i }));

      await waitFor(() => expect(mockedUpdate).toHaveBeenCalledTimes(1));
      expect(mockedUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { uid: "uid123" },
          body: expect.objectContaining({
            token: "tok456",
            password: "Secret123",
          }),
        }),
      );
    });

    it("navigates to /login with a success notice after successful submission", async () => {
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(screen.getByLabelText(/^confirm password$/i), "Secret123");
      await user.click(screen.getByRole("button", { name: /update password/i }));

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
      mockedUpdate.mockRejectedValue(new Error("network error"));
      const user = userEvent.setup();
      renderWithParams();

      await user.type(screen.getByLabelText(/^new password$/i), "Secret123");
      await user.type(screen.getByLabelText(/^confirm password$/i), "Secret123");
      await user.click(screen.getByRole("button", { name: /update password/i }));

      expect(await screen.findByRole("alert")).toBeInTheDocument();
    });
  });
});
