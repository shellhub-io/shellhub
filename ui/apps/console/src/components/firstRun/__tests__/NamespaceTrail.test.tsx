import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockUserAuth } from "@/tests/factories";
import { getConfig, defaultConfig } from "@/env";
import { ONBOARDING_TRIGGER } from "../survey/onboardingSurvey";
import NamespaceTrail from "../NamespaceTrail";

vi.mock("@/components/layout/SessionMenu", () => ({
  default: () => <div data-testid="session-menu" />,
}));

const SURVEY_API = "https://forms.example.test/api/v1/client/ws1";
const USER_ID = "user-123";

const survey = {
  id: "survey-1",
  type: "app",
  status: "inProgress",
  triggers: [{ actionClass: { name: ONBOARDING_TRIGGER } }],
  hiddenFields: { enabled: true, fieldIds: ["instance_type"] },
  blocks: [
    {
      elements: [
        {
          type: "contactInfo",
          id: "contact",
          headline: { default: "Contact information" },
          required: false,
          firstName: {
            show: true,
            required: false,
            placeholder: { default: "Name" },
          },
          lastName: { show: false, required: false },
          email: {
            show: true,
            required: false,
            placeholder: { default: "Email" },
          },
          phone: { show: false, required: false },
          company: { show: false, required: false },
        },
        {
          type: "multipleChoiceSingle",
          id: "role",
          headline: { default: "What's your role?" },
          required: true,
          choices: [{ id: "dev", label: { default: "Developer" } }],
        },
      ],
    },
  ],
};

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

beforeEach(() => {
  sent = [];
  localStorage.clear();
  mockGetConfig.mockReturnValue({
    ...defaultConfig,
    edition: "cloud",
    onboardingUrl: SURVEY_API,
  });
  seedAuthStore({
    userId: USER_ID,
    name: "Ana Maria Souza",
    email: "ana@example.com",
  });
  server.use(
    http.get("*/api/auth/user", () => HttpResponse.json(mockUserAuth())),
    http.get(`${SURVEY_API}/environment`, () =>
      HttpResponse.json({ data: { data: { surveys: [survey] } } }),
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
    expect(
      screen.queryByRole("textbox", { name: /namespace/i }),
    ).not.toBeInTheDocument();
  });

  it("sends the answers as cloud and moves on to creating a namespace", async () => {
    const user = userEvent.setup();
    renderTrail();

    await user.click(await screen.findByRole("radio", { name: /developer/i }));
    await user.click(screen.getByRole("button", { name: /^continue$/i }));

    expect(
      await screen.findByRole("button", { name: /back to the survey/i }),
    ).toBeInTheDocument();
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

  it("starts on the namespace step when this browser already sent a response", async () => {
    const user = userEvent.setup();
    const first = renderTrail();
    await user.click(await screen.findByRole("radio", { name: /developer/i }));
    await user.click(screen.getByRole("button", { name: /^continue$/i }));
    await screen.findByRole("button", { name: /back to the survey/i });
    first.unmount();

    renderTrail();

    expect(
      await screen.findByRole("button", { name: /back to the survey/i }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("radiogroup")).not.toBeInTheDocument();
  });

  it("updates the saved response instead of sending a second one", async () => {
    const user = userEvent.setup();
    const first = renderTrail();
    await user.click(await screen.findByRole("radio", { name: /developer/i }));
    await user.click(screen.getByRole("button", { name: /^continue$/i }));
    await screen.findByRole("button", { name: /back to the survey/i });
    first.unmount();

    renderTrail();
    await user.click(
      await screen.findByRole("button", { name: /back to the survey/i }),
    );
    expect(
      await screen.findByRole("radio", { name: /developer/i }),
    ).toBeChecked();
    await user.click(screen.getByRole("button", { name: /^continue$/i }));

    await waitFor(() =>
      expect(sent.map((s) => s.method)).toEqual(["POST", "PUT"]),
    );
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
