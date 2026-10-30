import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { delay, http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import {
  choiceElement,
  contactElement,
  mockSurveyPayload,
  surveyEnvironment,
} from "@/tests/onboardingSurvey";
import { getConfig, defaultConfig, type ClientConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import OnboardingSurveyGate from "../OnboardingSurveyGate";

vi.mock("@/components/layout/SessionMenu", () => ({
  default: () => <div data-testid="session-menu" />,
}));

const SURVEY_API = "https://forms.example.test/api/v1/client/ws1";

const survey = mockSurveyPayload({
  hiddenFields: {
    enabled: true,
    fieldIds: ["instance_type", "instance_domain"],
  },
  elements: [
    contactElement(),
    choiceElement("role", "What's your role?", [["dev", "Developer"]]),
  ],
});

const mockGetConfig = vi.mocked(getConfig);
let formbricks: { method: string; body?: unknown }[];
let recorded: unknown[];

function configure(overrides: Partial<ClientConfig>) {
  mockGetConfig.mockReturnValue({
    ...defaultConfig,
    edition: "cloud",
    onboardingUrl: SURVEY_API,
    ...overrides,
  });
}

function renderGate() {
  return render(
    <MemoryRouter initialEntries={["/devices"]}>
      <Routes>
        <Route element={<OnboardingSurveyGate />}>
          <Route path="/devices" element={<p>The devices page</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  formbricks = [];
  recorded = [];
  configure({});
  seedAuthStore({
    name: "Ana Maria Souza",
    email: "ana@example.com",
    onboardingSurvey: "pending",
  });
  server.use(
    http.get(`${SURVEY_API}/environment`, () => {
      formbricks.push({ method: "GET" });
      return HttpResponse.json(surveyEnvironment(survey));
    }),
    http.post(`${SURVEY_API}/responses`, async ({ request }) => {
      formbricks.push({ method: "POST", body: await request.json() });
      return HttpResponse.json({ data: { id: "response-1" } });
    }),
    http.put("*/api/users/onboarding-survey", async ({ request }) => {
      recorded.push(await request.json());
      return new HttpResponse(null, { status: 200 });
    }),
  );
});

describe("OnboardingSurveyGate", () => {
  it.each(["cloud", "community"] as const)(
    "on %s, shows a pending user the survey before the page they asked for",
    async (edition) => {
      configure({ edition });
      renderGate();

      expect(await screen.findByLabelText(/^name$/i)).toHaveValue(
        "Ana Maria Souza",
      );
      expect(screen.queryByText("The devices page")).not.toBeInTheDocument();
    },
  );

  it("says the survey is loading while the workspace has not answered yet", async () => {
    server.use(
      http.get(`${SURVEY_API}/environment`, () =>
        delay("infinite").then(() => HttpResponse.json({})),
      ),
    );
    renderGate();

    expect(await screen.findByText(/loading survey/i)).toBeInTheDocument();
    expect(screen.queryByText("The devices page")).not.toBeInTheDocument();
  });

  it("records the answer and moves on once the user answers", async () => {
    const user = userEvent.setup();
    renderGate();

    await user.click(await screen.findByRole("button", { name: /^next$/i }));
    await user.click(await screen.findByRole("radio", { name: /developer/i }));
    await user.click(screen.getByRole("button", { name: /^continue$/i }));

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
    expect(formbricks.filter((r) => r.method === "POST")).toEqual([
      {
        method: "POST",
        body: {
          surveyId: "survey-1",
          finished: true,
          data: {
            contact: ["Ana Maria Souza", "", "ana@example.com", "", ""],
            role: "Developer",
            instance_type: "cloud",
            instance_domain: window.location.hostname,
          },
        },
      },
    ]);
    expect(recorded).toEqual([{ outcome: "answered" }]);
    expect(useAuthStore.getState().onboardingSurvey).toBe("answered");
  });

  it("asks one question at a time", async () => {
    const user = userEvent.setup();
    renderGate();

    expect(await screen.findByText(/question 1 of 2/i)).toBeInTheDocument();
    expect(screen.queryByRole("radio")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^next$/i }));

    expect(screen.getByText(/question 2 of 2/i)).toBeInTheDocument();
    expect(
      screen.getByRole("radio", { name: /developer/i }),
    ).toBeInTheDocument();
    expect(screen.queryByLabelText(/^name$/i)).not.toBeInTheDocument();
  });

  it("keeps the answers when stepping back a question", async () => {
    const user = userEvent.setup();
    renderGate();
    const name = await screen.findByLabelText(/^name$/i);
    await user.clear(name);
    await user.type(name, "Ana");
    await user.click(screen.getByRole("button", { name: /^next$/i }));

    await user.click(screen.getByRole("button", { name: /^back$/i }));

    expect(screen.getByLabelText(/^name$/i)).toHaveValue("Ana");
  });

  it("shows the last question before a double click on Next can send the answers", async () => {
    const user = userEvent.setup();
    renderGate();
    await user.click(await screen.findByRole("button", { name: /^next$/i }));
    await user.click(screen.getByRole("radio", { name: /developer/i }));
    await user.click(screen.getByRole("button", { name: /^back$/i }));

    await user.dblClick(screen.getByRole("button", { name: /^next$/i }));

    expect(await screen.findByText(/question 2 of 2/i)).toBeInTheDocument();
    expect(formbricks.filter((r) => r.method === "POST")).toEqual([]);
    expect(recorded).toEqual([]);
  });

  it("holds a required question until it is answered", async () => {
    server.use(
      http.get(`${SURVEY_API}/environment`, () =>
        HttpResponse.json(
          surveyEnvironment(
            mockSurveyPayload({
              elements: [
                choiceElement("role", "What's your role?", [
                  ["dev", "Developer"],
                ]),
                contactElement(),
              ],
            }),
          ),
        ),
      ),
    );
    const user = userEvent.setup();
    renderGate();

    await user.click(await screen.findByRole("button", { name: /^next$/i }));

    expect(await screen.findByText(/pick one/i)).toBeInTheDocument();
    expect(screen.getByText(/question 1 of 2/i)).toBeInTheDocument();
  });

  it("records the skip and moves on once the user skips from the first question", async () => {
    const user = userEvent.setup();
    renderGate();

    await user.click(
      await screen.findByRole("button", { name: /skip survey/i }),
    );

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
    expect(formbricks.filter((r) => r.method === "POST")).toHaveLength(0);
    expect(recorded).toEqual([{ outcome: "skipped" }]);
    expect(useAuthStore.getState().onboardingSurvey).toBe("skipped");
  });

  it("says the answers could not be sent and still lets the user skip", async () => {
    server.use(
      http.post(`${SURVEY_API}/responses`, () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderGate();

    await user.click(await screen.findByRole("button", { name: /^next$/i }));
    await user.click(await screen.findByRole("radio", { name: /developer/i }));
    await user.click(screen.getByRole("button", { name: /^continue$/i }));

    expect(
      await screen.findByText(/your answers could not be sent/i),
    ).toBeInTheDocument();
    expect(recorded).toEqual([]);

    await user.click(screen.getByRole("button", { name: /skip survey/i }));

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
    expect(recorded).toEqual([{ outcome: "skipped" }]);
  });

  it("reads the survey from a URL configured with a trailing slash", async () => {
    configure({ onboardingUrl: `${SURVEY_API}/` });
    renderGate();

    expect(await screen.findByLabelText(/^name$/i)).toBeInTheDocument();
  });

  it("lets a pending user through, recording nothing, when the survey workspace never answers", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    try {
      server.use(
        http.get(`${SURVEY_API}/environment`, () =>
          delay("infinite").then(() => HttpResponse.json({})),
        ),
      );
      renderGate();
      expect(screen.queryByText("The devices page")).not.toBeInTheDocument();

      await vi.advanceTimersByTimeAsync(5000);

      expect(await screen.findByText("The devices page")).toBeInTheDocument();
      expect(recorded).toEqual([]);
    } finally {
      vi.useRealTimers();
    }
  });

  it("lets a pending user through, recording nothing, when the workspace has no onboarding survey", async () => {
    server.use(
      http.get(`${SURVEY_API}/environment`, () =>
        HttpResponse.json(surveyEnvironment()),
      ),
    );
    renderGate();

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
    expect(recorded).toEqual([]);
  });

  it("moves on for this visit even when recording the outcome fails", async () => {
    server.use(
      http.put(
        "*/api/users/onboarding-survey",
        () => new HttpResponse(null, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderGate();

    await user.click(
      await screen.findByRole("button", { name: /skip survey/i }),
    );

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
  });

  it("lets a pending user through, recording nothing, when no survey is configured", async () => {
    configure({ onboardingUrl: "" });
    renderGate();

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
    expect(formbricks).toEqual([]);
    expect(recorded).toEqual([]);
    expect(useAuthStore.getState().onboardingSurvey).toBe("pending");
  });

  it("lets a pending user through, recording nothing, when the survey workspace fails", async () => {
    server.use(
      http.get(
        `${SURVEY_API}/environment`,
        () => new HttpResponse(null, { status: 500 }),
      ),
    );
    renderGate();

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
    expect(recorded).toEqual([]);
    expect(useAuthStore.getState().onboardingSurvey).toBe("pending");
  });

  it.each(["answered", "skipped"] as const)(
    "lets a user who %s through without asking the survey workspace",
    async (outcome) => {
      seedAuthStore({ onboardingSurvey: outcome });
      renderGate();

      expect(await screen.findByText("The devices page")).toBeInTheDocument();
      expect(formbricks).toEqual([]);
    },
  );

  it("lets an administrator signed in as a pending user through without asking the survey workspace", async () => {
    seedAuthStore({ onboardingSurvey: "pending", impersonated: true });
    renderGate();

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
    expect(formbricks).toEqual([]);
    expect(recorded).toEqual([]);
  });

  it("lets a pending user through on enterprise without asking the survey workspace", async () => {
    configure({ edition: "enterprise" });
    renderGate();

    expect(await screen.findByText("The devices page")).toBeInTheDocument();
    expect(formbricks).toEqual([]);
    expect(recorded).toEqual([]);
  });
});
