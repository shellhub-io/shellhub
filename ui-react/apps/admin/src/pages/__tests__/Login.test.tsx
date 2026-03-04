import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { AxiosError } from "axios";
import { useAuthStore } from "../../stores/authStore";
import Login from "../Login";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("../../api/client", () => ({
  default: { post: vi.fn() },
}));

vi.mock("../../api/auth", () => ({
  getAuthUser: vi.fn(),
  updateUser: vi.fn(),
  updatePassword: vi.fn(),
  deleteUser: vi.fn(),
  resendEmail: vi.fn(),
}));

import apiClient from "../../api/client";
const mockedPost = vi.mocked(apiClient.post);

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function makeAxiosError(
  status: number,
  headers: Record<string, string> = {},
): AxiosError {
  const error = new AxiosError("Request failed");
  error.response = {
    status,
    data: {},
    headers: headers as never,
    config: {} as never,
    statusText: "Error",
  };
  return error;
}

function renderLogin() {
  return render(
    <MemoryRouter>
      <Login />
    </MemoryRouter>,
  );
}

async function fillAndSubmit(
  username = "admin",
  password = "secret",
  user = userEvent.setup(),
) {
  await user.type(screen.getByLabelText(/username/i), username);
  await user.type(screen.getByLabelText(/password/i), password);
  await user.click(screen.getByRole("button", { name: /sign in/i }));
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  mockNavigate.mockReset();
  mockedPost.mockReset();
  useAuthStore.setState({
    token: null,
    user: null,
    userId: null,
    email: null,
    username: null,
    recoveryEmail: null,
    tenant: null,
    role: null,
    name: null,
    loading: false,
  });
});

afterEach(() => {
  vi.useRealTimers();
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("Login", () => {
  describe("form rendering", () => {
    it("renders username and password fields with a submit button", () => {
      renderLogin();
      expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /sign in/i }),
      ).toBeInTheDocument();
    });

    it("shows no error by default", () => {
      renderLogin();
      expect(
        screen.queryByRole("alert"),
      ).not.toBeInTheDocument();
    });
  });

  describe("successful login", () => {
    it("navigates to /dashboard on success", async () => {
      mockedPost.mockResolvedValue({
        data: { token: "jwt", user: "admin", id: "uid", email: "admin@test.com", tenant: "tenant-1", name: "Admin" },
        headers: {},
      });

      renderLogin();
      await fillAndSubmit();

      expect(mockNavigate).toHaveBeenCalledWith("/dashboard");
    });
  });

  describe("loading state", () => {
    it("shows Authenticating... and disables the button while the request is in flight", async () => {
      let resolveLogin!: () => void;
      mockedPost.mockReturnValue(
        new Promise((resolve) => {
          resolveLogin = () =>
            resolve({ data: { token: "t", user: "u", id: "i", email: "e@e.com", tenant: "t", name: "n" }, headers: {} });
        }),
      );

      renderLogin();
      await userEvent.type(screen.getByLabelText(/username/i), "admin");
      await userEvent.type(screen.getByLabelText(/password/i), "secret");

      const clickPromise = userEvent.click(
        screen.getByRole("button", { name: /sign in/i }),
      );

      await waitFor(() =>
        expect(screen.getByText(/authenticating/i)).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: /authenticating/i })).toBeDisabled();

      resolveLogin();
      await clickPromise;
    });
  });

  describe("error handling", () => {
    it("shows invalid credentials error on 401", async () => {
      mockedPost.mockRejectedValue(makeAxiosError(401));

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/invalid login credentials/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows rate-limit error on 429", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 60;
      mockedPost.mockRejectedValue(
        makeAxiosError(429, { "x-account-lockout": String(epoch) }),
      );

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/too many failed login attempts/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows generic server error on unexpected status codes", async () => {
      mockedPost.mockRejectedValue(makeAxiosError(500));

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/something went wrong on our end/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows generic error on non-axios errors", async () => {
      mockedPost.mockRejectedValue(new Error("Network error"));

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/something went wrong\./i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("clears the error when a new submit is attempted", async () => {
      mockedPost.mockRejectedValueOnce(makeAxiosError(401));
      mockedPost.mockResolvedValueOnce({
        data: { token: "jwt", user: "admin", id: "uid", email: "admin@test.com", tenant: "tenant-1", name: "Admin" },
        headers: {},
      });

      const user = userEvent.setup();
      renderLogin();

      await fillAndSubmit("admin", "wrong", user);
      expect(screen.getByText(/invalid login credentials/i)).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /sign in/i }));
      expect(
        screen.queryByText(/invalid login credentials/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("429 countdown", () => {
    // These tests use real timers with short lockout epochs to avoid the
    // fake-timers + user.click() deadlock (React scheduler uses setTimeout(0)).

    it("displays the remaining lockout time after the first interval tick", async () => {
      // 30 seconds from now — first tick should show "29 seconds"
      const epoch = Math.floor(Date.now() / 1000) + 30;
      mockedPost.mockRejectedValue(
        makeAxiosError(429, { "x-account-lockout": String(epoch) }),
      );

      renderLogin();
      await fillAndSubmit();

      await waitFor(() =>
        expect(
          screen.getByText(/too many failed login attempts/i),
        ).toBeInTheDocument(),
      );

      // Wait for the first 1-second interval tick
      await waitFor(
        () => expect(screen.getByText(/seconds/i)).toBeInTheDocument(),
        { timeout: 2000 },
      );
    });

    it("shows lockout-expired alert when the countdown reaches zero", async () => {
      // 1 second lockout so the test completes quickly
      const epoch = Math.floor(Date.now() / 1000) + 1;
      mockedPost.mockRejectedValue(
        makeAxiosError(429, { "x-account-lockout": String(epoch) }),
      );

      renderLogin();
      await fillAndSubmit();

      await waitFor(() =>
        expect(
          screen.getByText(/too many failed login attempts/i),
        ).toBeInTheDocument(),
      );

      // Wait for the countdown to expire and the success alert to appear
      await waitFor(
        () =>
          expect(
            screen.getByText(/your timeout has finished/i),
          ).toBeInTheDocument(),
        { timeout: 4000 },
      );

      expect(
        screen.queryByText(/too many failed login attempts/i),
      ).not.toBeInTheDocument();
    });
  });
});
