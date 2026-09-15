import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import {
  PENDING_DEVICE_CODE_KEY,
  hasPendingDeviceCode,
  setPendingDeviceCode,
} from "@/utils/navigation";
import type { Info } from "@/client/model";
import { mockUserAuth } from "@/tests/factories";
import { simulateBrowserTranslation } from "@/tests/simulateBrowserTranslation";
import Login from "../Login";
import { getConfig, defaultConfig } from "@/env";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
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

function setLoginError(status: number, headers?: Record<string, string>) {
  server.use(
    http.post("*/api/login", () =>
      HttpResponse.json({}, { status, headers }),
    ),
  );
}

describe("Login", () => {
  beforeEach(() => {
    vi.clearAllMocks();
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
    server.use(
      http.post("*/api/login", () =>
        HttpResponse.json(mockUserAuth({ token: "jwt" })),
      ),
      http.get("*/info", () =>
        HttpResponse.json(
          mockInfo({ authentication: { local: true, saml: false } }),
        ),
      ),
      http.get("*/api/user/saml/auth", () =>
        HttpResponse.json({ url: "https://idp.example.com/sso" }),
      ),
    );
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

      expect(mockNavigate).toHaveBeenCalledWith("/dashboard");
    });

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
      renderLogin();
      await fillAndSubmit();

      expect(mockNavigate).toHaveBeenCalledWith("/dashboard");
    });
  });

  describe("loading state", () => {
    it("shows Authenticating... and disables the button while the request is in flight", async () => {
      let resolveHandler!: () => void;
      server.use(
        http.post(
          "*/api/login",
          () =>
            new Promise<Response>((resolve) => {
              resolveHandler = () =>
                resolve(HttpResponse.json(mockUserAuth()));
            }),
        ),
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

      resolveHandler();
      await clickPromise;
    });

    it("marks the submit button aria-busy while the request is in flight (DS Button loading prop)", async () => {
      let resolveHandler!: () => void;
      server.use(
        http.post(
          "*/api/login",
          () =>
            new Promise<Response>((resolve) => {
              resolveHandler = () =>
                resolve(HttpResponse.json(mockUserAuth()));
            }),
        ),
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
      ).toHaveAttribute("aria-busy", "true");

      resolveHandler();
      await clickPromise;
    });
  });

  describe("error handling", () => {
    it("shows invalid credentials error on 401", async () => {
      setLoginError(401);

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/invalid login credentials/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("redirects to confirm-account with the trimmed username on 403", async () => {
      setLoginError(403);

      renderLogin();
      await fillAndSubmit("  admin  ", "secret");

      expect(mockNavigate).toHaveBeenCalledWith(
        "/confirm-account?username=admin",
      );
    });

    it("shows rate-limit error on 429", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 60;
      setLoginError(429, { "x-account-lockout": String(epoch) });

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/too many failed login attempts/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows generic server error on unexpected status codes", async () => {
      setLoginError(500);

      renderLogin();
      await fillAndSubmit();

      expect(
        screen.getByText(/something went wrong on our end/i),
      ).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("shows generic error on network errors", async () => {
      server.use(
        http.post("*/api/login", () => HttpResponse.error()),
      );

      renderLogin();
      await fillAndSubmit();

      expect(screen.getByText(/something went wrong\./i)).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("clears the error when a new submit is attempted", async () => {
      let callCount = 0;
      server.use(
        http.post("*/api/login", () => {
          callCount++;
          if (callCount === 1) {
            return HttpResponse.json({}, { status: 401 });
          }
          return HttpResponse.json(mockUserAuth({ token: "jwt" }));
        }),
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
    it("displays the remaining lockout time after the first interval tick", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 30;
      setLoginError(429, { "x-account-lockout": String(epoch) });

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
      setLoginError(429, { "x-account-lockout": String(epoch) });

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
      setLoginError(429, { "x-account-lockout": String(epoch) });

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
    function setInfo(auth: { local: boolean; saml: boolean }) {
      server.use(
        http.get("*/info", () =>
          HttpResponse.json(mockInfo({ authentication: auth })),
        ),
      );
    }

    it("does not show SSO button on community edition", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig });
      setInfo({ local: true, saml: true });

      renderLogin();

      await waitFor(() =>
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument(),
      );

      expect(screen.queryByTestId("sso-btn")).not.toBeInTheDocument();
    });

    it("does not show SSO button when saml is false", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      setInfo({ local: true, saml: false });

      renderLogin();

      await waitFor(() =>
        expect(screen.getByLabelText(/username/i)).toBeInTheDocument(),
      );

      expect(screen.queryByTestId("sso-btn")).not.toBeInTheDocument();
    });

    it.each(["enterprise", "cloud"] as const)(
      "shows SSO button when edition=%s and saml is true",
      async (edition) => {
        mockGetConfig.mockReturnValue({ ...defaultConfig, edition });
        setInfo({ local: true, saml: true });

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
        setInfo({ local: true, saml: true });

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
      setInfo({ local: true, saml: true });
      server.use(
        http.get("*/api/user/saml/auth", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );

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
      setInfo({ local: false, saml: true });

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
      setPendingDeviceCode("WXYZ2K7Q");

      renderLogin();
      await fillAndSubmit();

      expect(mockNavigate).toHaveBeenCalledWith("/accept-device?code=WXYZ2K7Q");
      expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
    });

    it("prefers an explicit redirect over the pending code", async () => {
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
      server.use(
        http.post("*/api/login", () =>
          HttpResponse.json(mockUserAuth({ token: "jwt" }), {
            status: 401,
            headers: { "x-mfa-token": "mfa-temp" },
          }),
        ),
      );
      setPendingDeviceCode("WXYZ2K7Q");

      renderLogin();
      await fillAndSubmit();

      expect(mockNavigate).toHaveBeenCalledWith("/mfa-login");
      expect(hasPendingDeviceCode()).toBe(true);
    });
  });
});
