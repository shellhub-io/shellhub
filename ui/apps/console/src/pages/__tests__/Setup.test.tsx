import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import Setup from "../Setup";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/client", () => ({
  setup: vi.fn(),
}));

const mockLoginWithToken = vi.hoisted(() => vi.fn());

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (
    selector: (s: { loginWithToken: typeof mockLoginWithToken }) => unknown,
  ) => selector({ loginWithToken: mockLoginWithToken }),
}));

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: vi.fn(() => actual.getConfig()) };
});

import { setup as setupSdk } from "@/client";
import { getConfig, defaultConfig } from "@/env";

const mockedSetup = vi.mocked(setupSdk);
const mockedGetConfig = vi.mocked(getConfig);

type SdkResponse<T = unknown> = {
  data: T;
  request: Request;
  response: Response;
};

function mockSdkResponse<T>(data: T): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(),
  };
}

function makeSdkError(status: number) {
  return Object.assign(new Error("Request failed"), { status });
}

function renderSetup() {
  return render(
    <MemoryRouter>
      <Setup />
    </MemoryRouter>,
  );
}

/**
 * Fill all account fields with valid values. Does NOT submit.
 */
async function fillValidForm(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/^name$/i), "Alice Smith");
  await user.type(screen.getByLabelText(/^username$/i), "alice");
  await user.type(screen.getByLabelText(/^email$/i), "alice@example.com");
  await user.type(screen.getByLabelText(/^password$/i), "Secret123");
  await user.type(screen.getByLabelText(/^confirm password$/i), "Secret123");
}

afterEach(cleanup);

beforeEach(() => {
  mockNavigate.mockReset();
  mockedSetup.mockReset();
  mockLoginWithToken.mockReset();
  mockLoginWithToken.mockResolvedValue(undefined);
  mockedGetConfig.mockReturnValue({ ...defaultConfig });
});

describe("Setup", () => {
  describe("initial render", () => {
    it("renders all account form fields and the submit button", () => {
      renderSetup();
      expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^username$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^confirm password$/i)).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /complete setup/i }),
      ).toBeInTheDocument();
    });

    it("submit button is disabled when all fields are empty", () => {
      renderSetup();
      expect(
        screen.getByRole("button", { name: /complete setup/i }),
      ).toBeDisabled();
    });
  });

  describe("field validation — errors only after blur (onTouched mode)", () => {
    it("does not show name error before the field is touched", () => {
      renderSetup();
      expect(screen.queryByText(/name must be/i)).not.toBeInTheDocument();
    });

    it("shows name error after blurring an empty name field", async () => {
      const user = userEvent.setup();
      renderSetup();

      await user.click(screen.getByLabelText(/^name$/i));
      await user.tab();

      expect(await screen.findByText(/name must be/i)).toBeInTheDocument();
    });

    it("does not show username error before the field is touched", () => {
      renderSetup();
      expect(screen.queryByText(/username must be/i)).not.toBeInTheDocument();
    });

    it("shows username error after blurring with too-short value", async () => {
      const user = userEvent.setup();
      renderSetup();

      await user.type(screen.getByLabelText(/^username$/i), "ab");
      await user.tab();

      expect(await screen.findByText(/username must be/i)).toBeInTheDocument();
    });

    it("does not show email error before the field is touched", () => {
      renderSetup();
      expect(
        screen.queryByText(/enter a valid email/i),
      ).not.toBeInTheDocument();
    });

    it("shows email error after blurring with an invalid address", async () => {
      const user = userEvent.setup();
      renderSetup();

      await user.type(screen.getByLabelText(/^email$/i), "not-an-email");
      await user.tab();

      expect(
        await screen.findByText(/enter a valid email/i),
      ).toBeInTheDocument();
    });

    it("does not show password error before the field is touched", () => {
      renderSetup();
      expect(screen.queryByText(/password must be/i)).not.toBeInTheDocument();
    });

    it("shows password error after blurring an empty password field", async () => {
      const user = userEvent.setup();
      renderSetup();

      await user.click(screen.getByLabelText(/^password$/i));
      await user.tab();

      expect(await screen.findByText(/password must be/i)).toBeInTheDocument();
    });

    it("shows 'Passwords do not match' after blurring confirm password with a mismatched value", async () => {
      const user = userEvent.setup();
      renderSetup();

      await user.type(screen.getByLabelText(/^password$/i), "Secret123");
      await user.type(
        screen.getByLabelText(/^confirm password$/i),
        "Different",
      );
      await user.tab();

      expect(
        await screen.findByText(/passwords do not match/i),
      ).toBeInTheDocument();
    });
  });

  describe("submit gate", () => {
    it("enables the submit button only when all fields are valid", async () => {
      const user = userEvent.setup();
      renderSetup();

      const submit = screen.getByRole("button", { name: /complete setup/i });
      expect(submit).toBeDisabled();

      await fillValidForm(user);

      expect(submit).toBeEnabled();
    });
  });

  describe("successful submission", () => {
    // Setup responds with an authenticated session (auto-login); the token is what
    // the page hands to loginWithToken before redirecting into the app.
    const setupSuccess = () =>
      mockSdkResponse({ token: "jwt-token" }) as Awaited<
        ReturnType<typeof setupSdk>
      >;

    it("calls setup() with the correct payload and shows the success screen", async () => {
      mockedSetup.mockResolvedValue(setupSuccess());
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /complete setup/i }));

      await waitFor(() => expect(mockedSetup).toHaveBeenCalledTimes(1));
      expect(mockedSetup).toHaveBeenCalledWith(
        expect.objectContaining({
          body: {
            name: "Alice Smith",
            username: "alice",
            // Tests run with import.meta.env.DEV=true, so the namespace defaults to "dev".
            namespace: "dev",
            email: "alice@example.com",
            password: "Secret123",
          },
          throwOnError: true,
        }),
      );

      expect(await screen.findByText(/instance ready/i)).toBeInTheDocument();
    });

    it("logs in with the returned token", async () => {
      mockedSetup.mockResolvedValue(setupSuccess());
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /complete setup/i }));

      await waitFor(() =>
        expect(mockLoginWithToken).toHaveBeenCalledWith("jwt-token"),
      );
    });

    it("redirects to the app after 3 seconds on success", async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });

      mockedSetup.mockResolvedValue(setupSuccess());
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderSetup();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /complete setup/i }));

      await screen.findByText(/instance ready/i);

      vi.advanceTimersByTime(3000);

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith("/", {
          replace: true,
        }),
      );

      vi.useRealTimers();
    });

    it("routes to login with a notice when auto-login fails after setup", async () => {
      mockedSetup.mockResolvedValue(setupSuccess());
      mockLoginWithToken.mockRejectedValue(new Error("token login failed"));
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /complete setup/i }));

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith("/login", {
          replace: true,
          state: { notice: "Setup complete. Please sign in." },
        }),
      );
      // Setup succeeded, so the user must not see a failure error.
      expect(screen.queryByText(/an error occurred/i)).not.toBeInTheDocument();
    });

    it("routes to login when setup issues no token", async () => {
      mockedSetup.mockResolvedValue(
        mockSdkResponse({ token: "" }) as Awaited<ReturnType<typeof setupSdk>>,
      );
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /complete setup/i }));

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith("/login", {
          replace: true,
          state: { notice: "Setup complete. Please sign in." },
        }),
      );
      expect(mockLoginWithToken).not.toHaveBeenCalled();
    });
  });

  describe("error handling", () => {
    it("shows 'Setup has already been completed' on 409", async () => {
      mockedSetup.mockRejectedValue(makeSdkError(409));
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /complete setup/i }));

      expect(
        await screen.findByText(/setup has already been completed/i),
      ).toBeInTheDocument();
    });

    it("shows a generic error on unexpected server errors", async () => {
      mockedSetup.mockRejectedValue(makeSdkError(500));
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /complete setup/i }));

      expect(await screen.findByText(/an error occurred/i)).toBeInTheDocument();
    });
  });

  describe("two-step onboarding flow (when onboardingUrl is set)", () => {
    beforeEach(() => {
      mockedGetConfig.mockReturnValue({
        ...defaultConfig,
        onboardingUrl: "https://onboarding.example.com/survey",
      });
    });

    it("starts on the onboarding step and shows the survey iframe", () => {
      renderSetup();
      expect(screen.getByTitle(/onboarding survey/i)).toBeInTheDocument();
      expect(screen.queryByLabelText(/^name$/i)).not.toBeInTheDocument();
    });

    it("moves to the account step after Continue is clicked and survey is completed", async () => {
      const user = userEvent.setup();
      renderSetup();

      // Simulate the survey-completed postMessage from the iframe origin
      window.dispatchEvent(
        new MessageEvent("message", {
          data: "formbricksSurveyCompleted",
          origin: "https://onboarding.example.com",
        }),
      );

      await user.click(
        await screen.findByRole("button", { name: /continue/i }),
      );

      expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
    });

    it("shows a Back button on the account step that returns to onboarding", async () => {
      const user = userEvent.setup();
      renderSetup();

      window.dispatchEvent(
        new MessageEvent("message", {
          data: "formbricksSurveyCompleted",
          origin: "https://onboarding.example.com",
        }),
      );

      await user.click(
        await screen.findByRole("button", { name: /continue/i }),
      );

      expect(screen.getByRole("button", { name: /back/i })).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /back/i }));

      expect(screen.getByTitle(/onboarding survey/i)).toBeInTheDocument();
    });
  });
});
