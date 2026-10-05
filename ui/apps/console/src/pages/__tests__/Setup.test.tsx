import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { ONBOARDING_TRIGGER } from "@/components/firstRun/survey/onboardingSurvey";
import Setup from "../Setup";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const mockLoginWithToken = vi.hoisted(() => vi.fn());

vi.mock("@/stores/authStore", () => ({
  useAuthStore: Object.assign(
    (selector: (s: { loginWithToken: typeof mockLoginWithToken }) => unknown) =>
      selector({ loginWithToken: mockLoginWithToken }),
    {
      getState: () => ({
        token: null,
        logout: vi.fn(),
        setMfaToken: vi.fn(),
      }),
    },
  ),
}));
import { getConfig, defaultConfig } from "@/env";

const mockGetConfig = vi.mocked(getConfig);

function renderSetup() {
  return render(
    <MemoryRouter>
      <Setup />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

async function fillValidForm(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/^name$/i), "Alice Smith");
  await user.type(screen.getByLabelText(/^username$/i), "alice");
  await user.type(screen.getByLabelText(/^email$/i), "alice@example.com");
  await user.type(screen.getByLabelText(/^password$/i), "Secret123");
  await user.type(screen.getByLabelText(/^confirm password$/i), "Secret123");
}
beforeEach(() => {
  vi.clearAllMocks();
  mockLoginWithToken.mockResolvedValue(undefined);
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  server.use(
    http.post("*/api/setup", () => HttpResponse.json({ token: "jwt-token" })),
  );
});

describe("Setup", () => {
  describe("field validation — errors only after blur (onTouched mode)", () => {
    it("shows name error only after blurring an empty name field", async () => {
      const user = userEvent.setup();
      renderSetup();

      expect(screen.queryByText(/name must be/i)).not.toBeInTheDocument();

      await user.click(screen.getByLabelText(/^name$/i));
      await user.tab();

      expect(await screen.findByText(/name must be/i)).toBeInTheDocument();
    });

    it("shows username error only after blurring with a too-short value", async () => {
      const user = userEvent.setup();
      renderSetup();

      expect(screen.queryByText(/username must be/i)).not.toBeInTheDocument();

      await user.type(screen.getByLabelText(/^username$/i), "ab");
      await user.tab();

      expect(await screen.findByText(/username must be/i)).toBeInTheDocument();
    });

    it("shows email error only after blurring with an invalid address", async () => {
      const user = userEvent.setup();
      renderSetup();

      expect(
        screen.queryByText(/enter a valid email/i),
      ).not.toBeInTheDocument();

      await user.type(screen.getByLabelText(/^email$/i), "not-an-email");
      await user.tab();

      expect(
        await screen.findByText(/enter a valid email/i),
      ).toBeInTheDocument();
    });

    it("shows password error only after blurring an empty password field", async () => {
      const user = userEvent.setup();
      renderSetup();

      expect(screen.queryByText(/password must be/i)).not.toBeInTheDocument();

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

      const submit = screen.getByRole("button", {
        name: /create and continue/i,
      });
      expect(submit).toBeDisabled();

      await fillValidForm(user);

      expect(submit).toBeEnabled();
    });
  });

  describe("successful submission", () => {
    it("signs in with the returned token and continues on the dashboard trail", async () => {
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(
        screen.getByRole("button", { name: /create and continue/i }),
      );

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith("/dashboard", {
          replace: true,
          state: { firstRun: { fromSetup: true, survey: false } },
        }),
      );
      expect(mockLoginWithToken).toHaveBeenCalledWith("jwt-token");
    });

    it("lays out what comes after setup as the rest of the trail", () => {
      renderSetup();

      expect(
        screen.getByRole("heading", { name: /get your first shell/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByText(/create your account and namespace/i),
      ).toBeInTheDocument();
      expect(screen.getByText(/install the agent/i)).toBeInTheDocument();
      expect(screen.getByText(/open a shell on it/i)).toBeInTheDocument();
    });

    it("routes to login with a notice when auto-login fails after setup", async () => {
      mockLoginWithToken.mockRejectedValue(new Error("token login failed"));
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(
        screen.getByRole("button", { name: /create and continue/i }),
      );

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith("/login", {
          replace: true,
          state: { notice: "Setup complete. Please sign in." },
        }),
      );
      expect(screen.queryByText(/an error occurred/i)).not.toBeInTheDocument();
    });

    it("routes to login when setup issues no token", async () => {
      server.use(
        http.post("*/api/setup", () => HttpResponse.json({ token: "" })),
      );
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(
        screen.getByRole("button", { name: /create and continue/i }),
      );

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
      server.use(
        http.post("*/api/setup", () => HttpResponse.json({}, { status: 409 })),
      );
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(
        screen.getByRole("button", { name: /create and continue/i }),
      );

      expect(
        await screen.findByText(/setup has already been completed/i),
      ).toBeInTheDocument();
    });

    it("shows a generic error on unexpected server errors", async () => {
      server.use(
        http.post("*/api/setup", () => HttpResponse.json({}, { status: 500 })),
      );
      const user = userEvent.setup();
      renderSetup();

      await fillValidForm(user);
      await user.click(
        screen.getByRole("button", { name: /create and continue/i }),
      );

      expect(await screen.findByText(/an error occurred/i)).toBeInTheDocument();
    });
  });

  describe("onboarding survey (when onboardingUrl is set)", () => {
    const surveyApi = "https://forms.example.test/api/v1/client/ws1";
    let sent: { method: string; url: string; body: unknown }[];

    const survey = {
      id: "survey-1",
      type: "app",
      status: "inProgress",
      triggers: [{ actionClass: { name: ONBOARDING_TRIGGER } }],
      hiddenFields: { enabled: true, fieldIds: ["instance_domain"] },
      blocks: [
        {
          elements: [
            {
              type: "multipleChoiceSingle",
              id: "role",
              headline: { default: "What is your role?" },
              required: true,
              choices: [
                { id: "dev", label: { default: "Developer" } },
                { id: "ops", label: { default: "Operator" } },
              ],
            },
          ],
        },
      ],
    };

    function serveSurvey(environment: () => Response) {
      server.use(
        http.get(`${surveyApi}/environment`, environment),
        http.post(`${surveyApi}/responses`, async ({ request }) => {
          sent.push({
            method: "POST",
            url: request.url,
            body: await request.json(),
          });
          return HttpResponse.json({ data: { id: "response-1" } });
        }),
        http.put(`${surveyApi}/responses/:id`, async ({ request }) => {
          sent.push({
            method: "PUT",
            url: request.url,
            body: await request.json(),
          });
          return HttpResponse.json({ data: {} });
        }),
      );
    }

    beforeEach(() => {
      sent = [];
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        onboardingUrl: surveyApi,
      });
      serveSurvey(() =>
        HttpResponse.json({ data: { data: { surveys: [survey] } } }),
      );
    });

    async function answerAndContinue(
      user: ReturnType<typeof userEvent.setup>,
      choice: RegExp,
    ) {
      await user.click(await screen.findByRole("radio", { name: choice }));
      await user.click(screen.getByRole("button", { name: /^continue$/i }));
      await screen.findByLabelText(/^name$/i);
    }

    it("starts on the survey, rendered from the Formbricks workspace", async () => {
      renderSetup();

      expect(
        await screen.findByRole("radiogroup", { name: /what is your role/i }),
      ).toBeInTheDocument();
      expect(screen.queryByLabelText(/^name$/i)).not.toBeInTheDocument();
    });

    it("holds the user on the survey until a required question is answered", async () => {
      const user = userEvent.setup();
      renderSetup();

      await user.click(
        await screen.findByRole("button", { name: /^continue$/i }),
      );

      expect(await screen.findByText(/pick one/i)).toBeInTheDocument();
      expect(sent).toEqual([]);
    });

    it("sends the answers with the instance domain, then moves to the account step", async () => {
      const user = userEvent.setup();
      renderSetup();

      await answerAndContinue(user, /developer/i);

      expect(sent).toEqual([
        {
          method: "POST",
          url: `${surveyApi}/responses`,
          body: {
            surveyId: "survey-1",
            finished: true,
            data: {
              role: "Developer",
              instance_domain: window.location.hostname,
            },
          },
        },
      ]);
    });

    it("updates the response already sent when the user goes back and changes an answer", async () => {
      const user = userEvent.setup();
      renderSetup();

      await answerAndContinue(user, /developer/i);
      await user.click(screen.getByRole("button", { name: /back/i }));

      expect(
        await screen.findByRole("radio", { name: /developer/i }),
      ).toBeChecked();

      await answerAndContinue(user, /operator/i);

      expect(sent.map((s) => [s.method, s.url])).toEqual([
        ["POST", `${surveyApi}/responses`],
        ["PUT", `${surveyApi}/responses/response-1`],
      ]);
      expect(sent[1].body).toEqual({
        finished: true,
        data: { role: "Operator", instance_domain: window.location.hostname },
      });
    });

    it("tells the dashboard the survey is behind the user", async () => {
      const user = userEvent.setup();
      renderSetup();

      await answerAndContinue(user, /developer/i);
      await fillValidForm(user);
      await user.click(
        screen.getByRole("button", { name: /create and continue/i }),
      );

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith("/dashboard", {
          replace: true,
          state: { firstRun: { fromSetup: true, survey: true } },
        }),
      );
    });

    it("offers to skip when the answers cannot be sent", async () => {
      server.use(
        http.post(`${surveyApi}/responses`, () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      renderSetup();

      await user.click(
        await screen.findByRole("radio", { name: /developer/i }),
      );
      await user.click(screen.getByRole("button", { name: /^continue$/i }));

      expect(
        await screen.findByText(/your answers could not be sent/i),
      ).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /skip survey/i }));

      expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
    });

    it.each([
      [
        "the workspace cannot be reached",
        () => HttpResponse.json({}, { status: 503 }),
      ],
      [
        "the workspace has no onboarding survey",
        () => HttpResponse.json({ data: { data: { surveys: [] } } }),
      ],
    ])(
      "goes straight to the account step when %s",
      async (_case, environment) => {
        serveSurvey(environment);
        renderSetup();

        expect(await screen.findByLabelText(/^name$/i)).toBeInTheDocument();
        expect(screen.queryByRole("radiogroup")).not.toBeInTheDocument();
      },
    );
  });
});
