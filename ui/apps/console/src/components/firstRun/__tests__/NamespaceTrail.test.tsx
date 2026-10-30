import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockUserAuth } from "@/tests/factories";
import { mockSurveyPayload, surveyEnvironment } from "@/tests/onboardingSurvey";
import { getConfig, defaultConfig } from "@/env";
import NamespaceTrail from "../NamespaceTrail";

vi.mock("@/components/layout/SessionMenu", () => ({
  default: () => <div data-testid="session-menu" />,
}));

const SURVEY_API = "https://forms.example.test/api/v1/client/ws1";

const mockGetConfig = vi.mocked(getConfig);
let surveyRequests: number;

beforeEach(() => {
  surveyRequests = 0;
  mockGetConfig.mockReturnValue({
    ...defaultConfig,
    edition: "cloud",
    onboardingUrl: SURVEY_API,
  });
  seedAuthStore({ onboardingSurvey: "pending" });
  server.use(
    http.get("*/api/auth/user", () => HttpResponse.json(mockUserAuth())),
    http.get(`${SURVEY_API}/environment`, () => {
      surveyRequests++;
      return HttpResponse.json(surveyEnvironment(mockSurveyPayload()));
    }),
  );
});

describe("NamespaceTrail", () => {
  it("opens on creating a namespace, leaving the survey to the sign-in gate", async () => {
    render(
      <MemoryRouter>
        <NamespaceTrail />
      </MemoryRouter>,
      { wrapper: createTestWrapper() },
    );

    expect(
      await screen.findByText(/a namespace holds your devices/i),
    ).toBeInTheDocument();
    expect(screen.queryByRole("radiogroup")).not.toBeInTheDocument();
    expect(surveyRequests).toBe(0);
  });
});
