import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import {
  PENDING_DEVICE_CODE_KEY,
  hasPendingDeviceCode,
  setPendingDeviceCode,
} from "@/utils/navigation";
import type { Info, UserAuth } from "@/client";
import { mockUserAuth } from "@/tests/factories";
import { simulateBrowserTranslation } from "@/tests/simulateBrowserTranslation";
import Login from "../Login";
import { getConfig, defaultConfig } from "@/env";
import { mockSdkResponse, makeSdkError, type SdkResponse } from "@/tests/sdk";

const mockNavigate = vi.hoisted(() => vi.fn());
const mockLogin = vi.hoisted(() => vi.fn());
const mockGetInfo = vi.hoisted(() => vi.fn());
const mockGetSamlAuthUrl = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    login: mockLogin,
    getInfo: mockGetInfo,
    getSamlAuthUrl: mockGetSamlAuthUrl,
  };
});

const mockGetConfig = vi.mocked(getConfig);

function mockInfo(overrides: Partial<Info> = {}): Info {
  return {
    version: "0.0.0",
    endpoints: null,
    setup: true,
    authentication: { local: true, saml: false },
    ...overrides,
  };
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

describe("Login", () => {
  beforeEach(() => {
    mockNavigate.mockReset();
    mockLogin.mockReset();
    mockGetSamlAuthUrl.mockReset();
    mockGetInfo.mockResolvedValue(
      mockSdkResponse(
        mockInfo({ authentication: { local: true, saml: false } }),
      ),
    );
    mockGetConfig.mockReturnValue({ ...defaultConfig });
    localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
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

      expect(mockLogin).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { username: "admin", password: "secret" },
        }),
      );
    });

    it("does not trim password", async () => {
      renderLogin();
      await fillAndSubmit("admin", "  secret  ");

      expect(mockLogin).toHaveBeenCalledWith(
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
      mockLogin.mockResolvedValue(
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
      mockLogin.mockReturnValue(
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
      mockLogin.mockReturnValue(
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
      mockLogin.mockRejectedValue(makeSdkError(401));

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/invalid login credentials/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("redirects to confirm-account with the trimmed username on 403", async () => {
      mockLogin.mockRejectedValue(makeSdkError(403));

      renderLogin();
      await fillAndSubmit("  admin  ", "secret");

      expect(mockNavigate).toHaveBeenCalledWith(
        "/confirm-account?username=admin",
      );
    });

    it("shows rate-limit error on 429", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 60;
      mockLogin.mockRejectedValue(
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
      mockLogin.mockRejectedValue(makeSdkError(500));

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/something went wrong on our end/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows generic error on non-axios errors", async () => {
      mockLogin.mockRejectedValue(new Error("Network error"));

      renderLogin();
      await fillAndSubmit();

      expect(screen.getByText(/something went wrong\./i)).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("clears the error when a new submit is attempted", async () => {
      mockLogin.mockRejectedValueOnce(makeSdkError(401));
      mockLogin.mockResolvedValueOnce(
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
      mockLogin.mockRejectedValue(
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
      mockLogin.mockRejectedValue(
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

  describe("under a browser-translated DOM", () => {
    it("keeps updating the lockout countdown", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 30;
      mockLogin.mockRejectedValue(
        makeSdkError(429, { "x-account-lockout": String(epoch) }),
      );

      const { container } = renderLogin();
      await fillAndSubmit();
      await screen.findByText(/too many failed login attempts/i);
      simulateBrowserTranslation(container);

      await waitFor(
        () => expect(screen.getByText(/seconds/i)).toBeInTheDocument(),
        { timeout: 2000 },
      );
    });
  });

  describe("SSO / SAML button", () => {
    it("does not show SSO button on community edition", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig });
      mockGetInfo.mockResolvedValue(
        mockSdkResponse(
          mockInfo({ authentication: { local: true, saml: true } }),
        ),
      );

      renderLogin();

      await waitFor(() => expect(mockGetInfo).toHaveBeenCalled());

      expect(screen.queryByTestId("sso-btn")).not.toBeInTheDocument();
    });

    it("does not show SSO button when saml is false", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      mockGetInfo.mockResolvedValue(
        mockSdkResponse(
          mockInfo({ authentication: { local: true, saml: false } }),
        ),
      );

      renderLogin();

      await waitFor(() => expect(mockGetInfo).toHaveBeenCalled());

      expect(screen.queryByTestId("sso-btn")).not.toBeInTheDocument();
    });

    it.each(["enterprise", "cloud"] as const)(
      "shows SSO button when edition=%s and saml is true",
      async (edition) => {
        mockGetConfig.mockReturnValue({ ...defaultConfig, edition });
        mockGetInfo.mockResolvedValue(
          mockSdkResponse(
            mockInfo({ authentication: { local: true, saml: true } }),
          ),
        );

        renderLogin();

        await waitFor(() =>
          expect(screen.getByTestId("sso-btn")).toBeInTheDocument(),
        );
      },
    );

    it("redirects to SSO URL when SSO button is clicked", async () => {
      const originalLocation = window.location;
      Object.defineProperty(window, "location", {
        writable: true,
        value: { ...originalLocation, replace: vi.fn() },
      });

      try {
        mockGetConfig.mockReturnValue({
          ...defaultConfig,
          edition: "enterprise",
        });
        mockGetInfo.mockResolvedValue(
          mockSdkResponse(
            mockInfo({ authentication: { local: true, saml: true } }),
          ),
        );
        mockGetSamlAuthUrl.mockResolvedValue(
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
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      mockGetInfo.mockResolvedValue(
        mockSdkResponse(
          mockInfo({ authentication: { local: true, saml: true } }),
        ),
      );
      mockGetSamlAuthUrl.mockRejectedValue(new Error("Network error"));

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
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      mockGetInfo.mockResolvedValue(
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

  describe("pending device code", () => {
    it("redirects to /accept-device when a pending code exists and no explicit redirect", async () => {
      mockLogin.mockResolvedValue(
        mockSdkResponse(mockUserAuth({ token: "jwt" })),
      );
      setPendingDeviceCode("WXYZ2K7Q");

      renderLogin();
      await fillAndSubmit();

      expect(mockNavigate).toHaveBeenCalledWith("/accept-device?code=WXYZ2K7Q");
      expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
    });

    it("prefers an explicit redirect over the pending code", async () => {
      mockLogin.mockResolvedValue(
        mockSdkResponse(mockUserAuth({ token: "jwt" })),
      );
      setPendingDeviceCode("WXYZ2K7Q");

      render(
        <MemoryRouter initialEntries={["/login?redirect=%2Fdevices"]}>
          <Login />
        </MemoryRouter>,
      );
      await fillAndSubmit();

      expect(mockNavigate).toHaveBeenCalledWith("/devices");
      expect(hasPendingDeviceCode()).toBe(true);
    });

    it("does not consume the code when MFA is required", async () => {
      mockLogin.mockImplementation(async () => {
        useAuthStore.setState({ mfaToken: "mfa-temp" });
        return mockSdkResponse(mockUserAuth({ token: "jwt" }));
      });
      setPendingDeviceCode("WXYZ2K7Q");

      renderLogin();
      await fillAndSubmit();

      expect(mockNavigate).toHaveBeenCalledWith("/mfa-login");
      expect(hasPendingDeviceCode()).toBe(true);
    });
  });
});
