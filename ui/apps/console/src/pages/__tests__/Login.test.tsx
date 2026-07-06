import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import type { UserAuth, Info } from "@/client";
import Login from "../Login";

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

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: vi.fn(() => actual.getConfig()) };
});

import {
  login as loginSdk,
  getInfo as getInfoSdk,
  getSamlAuthUrl as getSamlAuthUrlSdk,
} from "@/client";
import { getConfig, defaultConfig } from "@/env";

const mockedLogin = vi.mocked(loginSdk);
const mockedGetInfo = vi.mocked(getInfoSdk);
const mockedGetSamlAuthUrl = vi.mocked(getSamlAuthUrlSdk);
const mockedGetConfig = vi.mocked(getConfig);

type SdkResponse<T = unknown> = {
  data: T;
  request: Request;
  response: Response;
};

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

function mockSdkResponse<T>(
  data: T,
  headers: HeadersInit = {},
): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(null, { headers }),
  };
}

function makeSdkError(status: number, headers?: Record<string, string>) {
  const headerObj = new Headers(headers);
  return Object.assign(new Error("Request failed"), {
    status,
    headers: headerObj,
  });
}

function renderLogin() {
  return render(
    <MemoryRouter>
      <Login />
    </MemoryRouter>,
  );
}

// With RHF onTouched mode, fields must be blurred before validation runs.
// user.tab() after typing each field triggers the blur that enables the button.
async function fillAndSubmit(
  username = "admin",
  password = "secret",
  user = userEvent.setup(),
) {
  await user.type(screen.getByLabelText(/username/i), username);
  await user.tab();
  await user.type(screen.getByLabelText(/^password$/i), password);
  await user.tab();
  await user.click(screen.getByRole("button", { name: /sign in/i }));
}

afterEach(cleanup);

beforeEach(() => {
  mockNavigate.mockReset();
  mockedLogin.mockReset();
  mockedGetSamlAuthUrl.mockReset();
  mockedGetInfo.mockResolvedValue(
    mockSdkResponse(mockInfo({ authentication: { local: true, saml: false } })),
  );
  mockedGetConfig.mockReturnValue({ ...defaultConfig });
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

describe("Login", () => {
  describe("form rendering", () => {
    it("renders username and password fields with a submit button", () => {
      renderLogin();
      expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /sign in/i }),
      ).toBeInTheDocument();
    });

    it("shows no error by default", () => {
      renderLogin();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });

    it("trims username before submitting", async () => {
      renderLogin();
      await fillAndSubmit("  admin  ", "secret");

      expect(mockedLogin).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { username: "admin", password: "secret" },
        }),
      );
    });

    it("does not trim password", async () => {
      renderLogin();
      await fillAndSubmit("admin", "  secret  ");

      expect(mockedLogin).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { username: "admin", password: "  secret  " },
        }),
      );
    });

    // RHF onTouched: blurring an invalid field surfaces a per-field error
    // message inline. The old useState implementation never showed field errors.
    it("shows a field error on the username field after blur when empty", async () => {
      const user = userEvent.setup();
      renderLogin();

      await user.type(screen.getByLabelText(/username/i), "admin");
      await user.clear(screen.getByLabelText(/username/i));
      await user.tab();

      await waitFor(() =>
        expect(screen.getByText(/is required/i)).toBeInTheDocument(),
      );
    });

    // With RHF onTouched, validation runs after blur. The button is disabled
    // until both fields have been touched and are valid. Tab away from each
    // field to trigger blur-time validation before asserting enabled state.
    it("disables the submit button when username or password is empty", async () => {
      const user = userEvent.setup();
      renderLogin();

      const submitButton = screen.getByRole("button", { name: /sign in/i });
      expect(submitButton).toBeDisabled();

      await user.type(screen.getByLabelText(/username/i), "admin");
      await user.tab();
      expect(submitButton).toBeDisabled();

      await user.type(screen.getByLabelText(/^password$/i), "secret");
      await user.tab();
      await waitFor(() => expect(submitButton).toBeEnabled());

      await user.clear(screen.getByLabelText(/username/i));
      await user.tab();
      await waitFor(() => expect(submitButton).toBeDisabled());
    });
  });

  describe("successful login", () => {
    it("navigates to /dashboard on success", async () => {
      mockedLogin.mockResolvedValue(
        mockSdkResponse(mockUserAuth({ token: "jwt" })),
      );

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
      await userEvent.tab();
      await userEvent.type(screen.getByLabelText(/^password$/i), "secret");
      await userEvent.tab();

      const clickPromise = userEvent.click(
        screen.getByRole("button", { name: /sign in/i }),
      );

      await waitFor(() =>
        expect(screen.getByText(/authenticating/i)).toBeInTheDocument(),
      );
      expect(
        screen.getByRole("button", { name: /authenticating/i }),
      ).toBeDisabled();

      resolveLogin();
      await clickPromise;
    });

    it("marks the submit button aria-busy while the request is in flight (DS Button loading prop)", async () => {
      let resolveLogin!: () => void;
      mockedLogin.mockReturnValue(
        new Promise<SdkResponse<UserAuth>>((resolve) => {
          resolveLogin = () => resolve(mockSdkResponse(mockUserAuth()));
        }),
      );

      renderLogin();
      await userEvent.type(screen.getByLabelText(/username/i), "admin");
      await userEvent.tab();
      await userEvent.type(screen.getByLabelText(/^password$/i), "secret");
      await userEvent.tab();

      const clickPromise = userEvent.click(
        screen.getByRole("button", { name: /sign in/i }),
      );

      await waitFor(() =>
        expect(screen.getByText(/authenticating/i)).toBeInTheDocument(),
      );

      // DS Button sets aria-busy="true" on the button element when loading=true.
      expect(
        screen.getByRole("button", { name: /authenticating/i }),
      ).toHaveAttribute("aria-busy", "true");

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

    it("redirects to confirm-account with the trimmed username on 403", async () => {
      mockedLogin.mockRejectedValue(makeSdkError(403));

      renderLogin();
      await fillAndSubmit("  admin  ", "secret");

      expect(mockNavigate).toHaveBeenCalledWith(
        "/confirm-account?username=admin",
      );
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

      expect(screen.getByText(/something went wrong\./i)).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("clears the error when a new submit is attempted", async () => {
      mockedLogin.mockRejectedValueOnce(makeSdkError(401));
      mockedLogin.mockResolvedValueOnce(
        mockSdkResponse(mockUserAuth({ token: "jwt" })),
      );

      const user = userEvent.setup();
      renderLogin();

      await fillAndSubmit("admin", "wrong", user);
      expect(
        screen.getByText(/invalid login credentials/i),
      ).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /sign in/i }));
      expect(
        screen.queryByText(/invalid login credentials/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("429 countdown", () => {
    // Real timers: fake timers + user.click() deadlock (React scheduler uses setTimeout(0)).

    it("displays the remaining lockout time after the first interval tick", async () => {
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

      await waitFor(
        () => expect(screen.getByText(/seconds/i)).toBeInTheDocument(),
        { timeout: 2000 },
      );
    });

    it("shows lockout-expired alert when the countdown reaches zero", async () => {
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
      mockedGetConfig.mockReturnValue({ ...defaultConfig });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(
          mockInfo({ authentication: { local: true, saml: true } }),
        ),
      );

      renderLogin();

      await waitFor(() => expect(mockedGetInfo).toHaveBeenCalled());

      expect(screen.queryByTestId("sso-btn")).not.toBeInTheDocument();
    });

    it("does not show SSO button when enterprise but saml is false", async () => {
      mockedGetConfig.mockReturnValue({ ...defaultConfig, enterprise: true });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(
          mockInfo({ authentication: { local: true, saml: false } }),
        ),
      );

      renderLogin();

      await waitFor(() => expect(mockedGetInfo).toHaveBeenCalled());

      expect(screen.queryByTestId("sso-btn")).not.toBeInTheDocument();
    });

    it("shows SSO button when enterprise and saml is true", async () => {
      mockedGetConfig.mockReturnValue({ ...defaultConfig, enterprise: true });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(
          mockInfo({ authentication: { local: true, saml: true } }),
        ),
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
        mockedGetConfig.mockReturnValue({ ...defaultConfig, enterprise: true });
        mockedGetInfo.mockResolvedValue(
          mockSdkResponse(
            mockInfo({ authentication: { local: true, saml: true } }),
          ),
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
        Object.defineProperty(window, "location", {
          writable: true,
          value: originalLocation,
        });
      }
    });

    it("shows error when SSO URL fetch fails", async () => {
      mockedGetConfig.mockReturnValue({ ...defaultConfig, enterprise: true });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(
          mockInfo({ authentication: { local: true, saml: true } }),
        ),
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
      mockedGetConfig.mockReturnValue({ ...defaultConfig, enterprise: true });
      mockedGetInfo.mockResolvedValue(
        mockSdkResponse(
          mockInfo({ authentication: { local: false, saml: true } }),
        ),
      );

      renderLogin();

      expect(screen.queryByLabelText(/username/i)).not.toBeInTheDocument();
      expect(screen.queryByLabelText(/^password$/i)).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /sign in/i }),
      ).not.toBeInTheDocument();

      await waitFor(() =>
        expect(screen.getByTestId("sso-btn")).toBeInTheDocument(),
      );
    });
  });
});
