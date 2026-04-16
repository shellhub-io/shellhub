import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import type { UserAuth } from "@/client";
import Login from "../Login";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/client", () => ({
  login: vi.fn(),
  getUserInfo: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
  authMfa: vi.fn(),
  mfaRecover: vi.fn(),
  requestResetMfa: vi.fn(),
  resetMfa: vi.fn(),
  resendEmail: vi.fn(),
}));

import { login as loginSdk } from "@/client";
const mockedLogin = vi.mocked(loginSdk);

type SdkResponse<T = unknown> = { data: T; request: Request; response: Response };

function mockUserAuth(overrides: Partial<UserAuth> = {}): UserAuth {
  return {
    token: "jwt-token",
    id: "uid",
    origin: "local",
    user: "admin",
    name: "Admin",
    email: "admin@test.com",
    recovery_email: "recovery@test.com",
    tenant: "tenant-1",
    role: "owner",
    mfa: false,
    admin: false,
    max_namespaces: -1,
    ...overrides,
  };
}

function mockSdkResponse<T>(data: T, headers: HeadersInit = {}): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(null, { headers }),
  };
}

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

/** Creates a mock SDK error with status and optional headers. */
function makeSdkError(
  status: number,
  headers?: Record<string, string>,
) {
  const headerObj = new Headers(headers);
  return Object.assign(new Error("Request failed"), { status, headers: headerObj });
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
  mockedLogin.mockReset();
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
      mockedLogin.mockResolvedValue(mockSdkResponse(mockUserAuth({ token: "jwt" })));

      renderLogin();
      await fillAndSubmit();

      expect(mockNavigate).toHaveBeenCalledWith("/dashboard");
    });
  });

  describe("loading state", () => {
    it("shows Authenticating... and disables the button while the request is in flight", async () => {
      let resolveLogin!: () => void;
      mockedLogin.mockReturnValue(
        new Promise<SdkResponse<UserAuth>>((resolve) => {
          resolveLogin = () => resolve(mockSdkResponse(mockUserAuth()));
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
      mockedLogin.mockRejectedValue(makeSdkError(401));

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/invalid login credentials/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows rate-limit error on 429", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 60;
      mockedLogin.mockRejectedValue(
        makeSdkError(429, { "x-account-lockout": String(epoch) }),
      );

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/too many failed login attempts/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows generic server error on unexpected status codes", async () => {
      mockedLogin.mockRejectedValue(makeSdkError(500));

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/something went wrong on our end/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows generic error on non-axios errors", async () => {
      mockedLogin.mockRejectedValue(new Error("Network error"));

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/something went wrong\./i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("clears the error when a new submit is attempted", async () => {
      mockedLogin.mockRejectedValueOnce(makeSdkError(401));
      mockedLogin.mockResolvedValueOnce(mockSdkResponse(mockUserAuth({ token: "jwt" })));

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
      mockedLogin.mockRejectedValue(
        makeSdkError(429, { "x-account-lockout": String(epoch) }),
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
      mockedLogin.mockRejectedValue(
        makeSdkError(429, { "x-account-lockout": String(epoch) }),
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
