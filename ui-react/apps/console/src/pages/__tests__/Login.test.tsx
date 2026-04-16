import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";
import type { UserAuth, Info } from "../../client";
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
  getInfo: vi.fn(),
  getSamlAuthUrl: vi.fn(),
}));

vi.mock("../../env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../env")>();
  return { ...actual, getConfig: vi.fn(() => actual.getConfig()) };
});

import { login as loginSdk, getInfo as getInfoSdk, getSamlAuthUrl as getSamlAuthUrlSdk } from "../../client";
import { getConfig } from "../../env";

const mockedLogin = vi.mocked(loginSdk);
const mockedGetInfo = vi.mocked(getInfoSdk);
const mockedGetSamlAuthUrl = vi.mocked(getSamlAuthUrlSdk);
const mockedGetConfig = vi.mocked(getConfig);

type SdkResponse<T = unknown> = { data: T; request: Request; response: Response };

function mockInfo(overrides: Partial<Info> = {}): Info {
  return {
    version: "0.0.0",
    endpoints: null,
    setup: true,
    authentication: { local: true, saml: false },
    ...overrides,
  };
}

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
  mockedGetSamlAuthUrl.mockReset();
  // Default: local=true, saml=false — community/non-enterprise baseline.
  mockedGetInfo.mockResolvedValue(
    mockSdkResponse(mockInfo({ authentication: { local: true, saml: false } })),
  );
  // Default: community edition (no enterprise/cloud flags).
  mockedGetConfig.mockReturnValue({
    version: "",
    enterprise: false,
    cloud: false,
    announcements: false,
    onboardingUrl: "",
  });
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

  describe("SSO / SAML button", () => {
    it("does not show SSO button when not enterprise", async () => {
      mockedGetConfig.mockReturnValue({
        version: "",
        enterprise: false,
        cloud: false,
        announcements: false,
        onboardingUrl: "",
      });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(mockInfo({ authentication: { local: true, saml: true } })),
      );

      renderLogin();

      await waitFor(() => expect(mockedGetInfo).toHaveBeenCalled());

      expect(screen.queryByTestId("sso-btn")).not.toBeInTheDocument();
    });

    it("does not show SSO button when enterprise but saml is false", async () => {
      mockedGetConfig.mockReturnValue({
        version: "",
        enterprise: true,
        cloud: false,
        announcements: false,
        onboardingUrl: "",
      });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(mockInfo({ authentication: { local: true, saml: false } })),
      );

      renderLogin();

      // Wait for getInfo to resolve and state to update.
      await waitFor(() => expect(mockedGetInfo).toHaveBeenCalled());

      expect(screen.queryByTestId("sso-btn")).not.toBeInTheDocument();
    });

    it("shows SSO button when enterprise and saml is true", async () => {
      mockedGetConfig.mockReturnValue({
        version: "",
        enterprise: true,
        cloud: false,
        announcements: false,
        onboardingUrl: "",
      });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(mockInfo({ authentication: { local: true, saml: true } })),
      );

      renderLogin();

      await waitFor(() =>
        expect(screen.getByTestId("sso-btn")).toBeInTheDocument(),
      );
    });

    it("redirects to SSO URL when SSO button is clicked", async () => {
      const originalLocation = window.location;
      Object.defineProperty(window, "location", {
        writable: true,
        value: { ...originalLocation, replace: vi.fn() },
      });

      try {
        mockedGetConfig.mockReturnValue({
          version: "",
          enterprise: true,
          cloud: false,
          announcements: false,
          onboardingUrl: "",
        });
        mockedGetInfo.mockResolvedValue(
          mockSdkResponse(mockInfo({ authentication: { local: true, saml: true } })),
        );
        mockedGetSamlAuthUrl.mockResolvedValue(
          mockSdkResponse({ url: "https://idp.example.com/sso" }),
        );

        renderLogin();

        const ssoBtn = await screen.findByTestId("sso-btn");
        await userEvent.click(ssoBtn);

        await waitFor(() =>
          expect(window.location.replace).toHaveBeenCalledWith(
            "https://idp.example.com/sso",
          ),
        );
      } finally {
        Object.defineProperty(window, "location", { writable: true, value: originalLocation });
      }
    });

    it("shows error when SSO URL fetch fails", async () => {
      mockedGetConfig.mockReturnValue({
        version: "",
        enterprise: true,
        cloud: false,
        announcements: false,
        onboardingUrl: "",
      });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(mockInfo({ authentication: { local: true, saml: true } })),
      );
      mockedGetSamlAuthUrl.mockRejectedValue(new Error("Network error"));

      renderLogin();

      const ssoBtn = await screen.findByTestId("sso-btn");
      await userEvent.click(ssoBtn);

      await waitFor(() =>
        expect(
          screen.getByText(/failed to retrieve sso login url/i),
        ).toBeInTheDocument(),
      );
    });

    it("hides the form entirely and shows SSO as the only option when local auth is disabled", async () => {
      mockedGetConfig.mockReturnValue({
        version: "",
        enterprise: true,
        cloud: false,
        announcements: false,
        onboardingUrl: "",
      });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(mockInfo({ authentication: { local: false, saml: true } })),
      );

      renderLogin();

      // Form must be absent immediately — the explicit showLocalForm guard
      // prevents it from appearing even during the getInfo loading window.
      expect(screen.queryByLabelText(/username/i)).not.toBeInTheDocument();
      expect(screen.queryByLabelText(/password/i)).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: /sign in/i })).not.toBeInTheDocument();

      // SSO button appears once getInfo resolves.
      await waitFor(() =>
        expect(screen.getByTestId("sso-btn")).toBeInTheDocument(),
      );
    });
  });
});
