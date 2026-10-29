import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockUserAuth } from "@/tests/factories";
import {
  choiceElement,
  contactElement,
  mockSurveyPayload,
  surveyEnvironment,
} from "@/tests/onboardingSurvey";
import { getConfig, defaultConfig } from "@/env";
import { saveResponseId } from "@/utils/savedSurvey";
import NamespaceTrail from "../NamespaceTrail";

vi.mock("@/components/layout/SessionMenu", () => ({
  default: () => <div data-testid="session-menu" />,
}));

const SURVEY_API = "https://forms.example.test/api/v1/client/ws1";

const survey = mockSurveyPayload({
  hiddenFields: { enabled: true, fieldIds: ["instance_type"] },
  elements: [
    contactElement(),
    choiceElement("role", "What's your role?", [["dev", "Developer"]]),
  ],
});

const mockGetConfig = vi.mocked(getConfig);
let sent: { method: string; body: unknown }[];

function renderTrail() {
  return render(
    <MemoryRouter>
      <NamespaceTrail />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

function storedEntries(): Map<string, string | null> {
  return new Map(
    Array.from({ length: localStorage.length }, (_, i) => {
      const key = localStorage.key(i) ?? "";
      return [key, localStorage.getItem(key)];
    }),
  );
}

async function answerAndContinue(user: ReturnType<typeof userEvent.setup>) {
  await user.click(await screen.findByRole("button", { name: /^next$/i }));
  await user.click(await screen.findByRole("radio", { name: /developer/i }));
  await user.click(screen.getByRole("button", { name: /^continue$/i }));
  await screen.findByRole("button", { name: /back to the survey/i });
}

beforeEach(() => {
  sent = [];
  localStorage.clear();
  mockGetConfig.mockReturnValue({
    ...defaultConfig,
    edition: "cloud",
    onboardingUrl: SURVEY_API,
  });
  seedAuthStore({
    userId: "user-123",
    name: "Ana Maria Souza",
    email: "ana@example.com",
  });
  server.use(
    http.get("*/api/auth/user", () => HttpResponse.json(mockUserAuth())),
    http.get(`${SURVEY_API}/environment`, () =>
      HttpResponse.json(surveyEnvironment(survey)),
    ),
    http.post(`${SURVEY_API}/responses`, async ({ request }) => {
      sent.push({ method: "POST", body: await request.json() });
      return HttpResponse.json({ data: { id: "response-1" } });
    }),
    http.put(`${SURVEY_API}/responses/:id`, async ({ request }) => {
      sent.push({ method: "PUT", body: await request.json() });
      return HttpResponse.json({ data: {} });
    }),
  );
});

describe("NamespaceTrail onboarding survey", () => {
  it("opens on the survey with the account's whole name and email filled in", async () => {
    renderTrail();

    expect(await screen.findByLabelText(/^name$/i)).toHaveValue(
      "Ana Maria Souza",
    );
    expect(screen.getByLabelText(/^email$/i)).toHaveValue("ana@example.com");
  });

  it("sends the answers as cloud and moves on to creating a namespace", async () => {
    const user = userEvent.setup();
    renderTrail();

    await answerAndContinue(user);

    expect(sent).toEqual([
      {
        method: "POST",
        body: {
          surveyId: "survey-1",
          finished: true,
          data: {
            contact: ["Ana Maria Souza", "", "ana@example.com", "", ""],
            role: "Developer",
            instance_type: "cloud",
          },
        },
      },
    ]);
  });

  it("opens on the namespace step when this browser already sent a response", async () => {
    const user = userEvent.setup();
    const first = renderTrail();
    await answerAndContinue(user);
    first.unmount();

    renderTrail();

    expect(
      await screen.findByRole("button", { name: /back to the survey/i }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("radiogroup")).not.toBeInTheDocument();
  });

  it("updates the response sent on an earlier visit instead of sending a second one", async () => {
    const user = userEvent.setup();
    const first = renderTrail();
    await answerAndContinue(user);
    first.unmount();

    renderTrail();
    await user.click(
      await screen.findByRole("button", { name: /back to the survey/i }),
    );
    await answerAndContinue(user);

    await waitFor(() =>
      expect(sent.map((s) => s.method)).toEqual(["POST", "PUT"]),
    );
  });

  it("adds only the response id to browser storage, never the answers", async () => {
    const user = userEvent.setup();
    renderTrail();
    await screen.findByLabelText(/^name$/i);
    const before = storedEntries();

    await answerAndContinue(user);

    const added = [...storedEntries()]
      .filter(([key, value]) => before.get(key) !== value)
      .map(([, value]) => value);
    expect(added).toEqual(["response-1"]);
  });

  it("sends a new response when the saved one no longer exists", async () => {
    saveResponseId("user-123", "survey-1", "deleted");
    server.use(
      http.put(`${SURVEY_API}/responses/:id`, async ({ request }) => {
        sent.push({ method: "PUT", body: await request.json() });
        return HttpResponse.json({}, { status: 404 });
      }),
    );
    const user = userEvent.setup();
    renderTrail();

    await user.click(
      await screen.findByRole("button", { name: /back to the survey/i }),
    );
    await answerAndContinue(user);

    expect(sent.map((s) => s.method)).toEqual(["PUT", "POST"]);
  });

  it("asks again when the trigger moved to another survey", async () => {
    saveResponseId("user-123", "old-survey", "response-0");
    renderTrail();

    expect(await screen.findByText(/question 1 of 2/i)).toBeInTheDocument();
  });

  it("asks one question at a time", async () => {
    const user = userEvent.setup();
    renderTrail();

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
    renderTrail();
    const name = await screen.findByLabelText(/^name$/i);
    await user.clear(name);
    await user.type(name, "Ana");
    await user.click(screen.getByRole("button", { name: /^next$/i }));

    await user.click(screen.getByRole("button", { name: /^back$/i }));

    expect(screen.getByLabelText(/^name$/i)).toHaveValue("Ana");
  });

  it("shows the last question before a double click on Next can send the answers", async () => {
    const user = userEvent.setup();
    renderTrail();
    await user.click(await screen.findByRole("button", { name: /^next$/i }));
    await user.click(screen.getByRole("radio", { name: /developer/i }));
    await user.click(screen.getByRole("button", { name: /^back$/i }));

    await user.dblClick(screen.getByRole("button", { name: /^next$/i }));

    expect(await screen.findByText(/question 2 of 2/i)).toBeInTheDocument();
    expect(sent).toEqual([]);
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
    renderTrail();

    await user.click(await screen.findByRole("button", { name: /^next$/i }));

    expect(await screen.findByText(/pick one/i)).toBeInTheDocument();
    expect(screen.getByText(/question 1 of 2/i)).toBeInTheDocument();
  });

  it("takes back the contact details sent earlier when the user turns anonymous", async () => {
    const user = userEvent.setup();
    renderTrail();
    await answerAndContinue(user);

    await user.click(
      screen.getByRole("button", { name: /back to the survey/i }),
    );
    await user.click(
      await screen.findByRole("checkbox", { name: /answer anonymously/i }),
    );
    await user.click(screen.getByRole("button", { name: /^next$/i }));
    await user.click(screen.getByRole("button", { name: /^continue$/i }));
    await screen.findByRole("button", { name: /back to the survey/i });

    expect(sent.map((s) => s.method)).toEqual(["POST", "PUT"]);
    expect(sent[1].body).toEqual({
      finished: true,
      data: {
        contact: ["", "", "", "", ""],
        role: "Developer",
        instance_type: "cloud",
      },
    });
  });

  it("skips the survey on enterprise", async () => {
    mockGetConfig.mockReturnValue({
      ...defaultConfig,
      edition: "enterprise",
      onboardingUrl: SURVEY_API,
    });
    renderTrail();

    expect(
      await screen.findByText(/a namespace holds your devices/i),
    ).toBeInTheDocument();
    expect(screen.queryByRole("radiogroup")).not.toBeInTheDocument();
  });
});
