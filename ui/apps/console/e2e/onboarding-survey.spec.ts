import { type Page, type Route, expect, test } from "@playwright/test";
import { getUserInfo } from "@/client";
import {
  choiceElement,
  contactElement,
  mockSurveyPayload,
  surveyEnvironment,
} from "@/tests/onboardingSurvey";
import { buildRequestContext, loginAs } from "./api";
import { isEnterprise } from "./env";
import { createTeam, signIn, signOut } from "./helpers";
import { password } from "./seed";

const surveyApi = "https://forms.e2e.test/api/v1/client/workspace";
const enterpriseReason = "enterprise never asks the onboarding survey";

const survey = mockSurveyPayload({
  elements: [
    contactElement(),
    choiceElement("role", "What's your role?", [["dev", "Developer"]]),
  ],
});

type FakeSurvey = {
  requests: string[];
  environment?: (route: Route) => Promise<void>;
};

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET, POST, PUT, OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type",
};

function answerJson(route: Route, body: unknown, status = 200) {
  return route.fulfill({ status, headers: corsHeaders, json: body });
}

async function serveFakeSurvey(
  page: Page,
  { environment }: Omit<FakeSurvey, "requests"> = {},
) {
  const fake: FakeSurvey = { requests: [], environment };

  await page.route("**/config.json", async (route) => {
    const response = await route.fetch();
    const config = (await response.json()) as Record<string, unknown>;
    await route.fulfill({
      response,
      json: { ...config, onboardingUrl: surveyApi },
    });
  });

  await page.route(`${surveyApi}/**`, async (route) => {
    const request = route.request();
    if (request.method() === "OPTIONS") {
      await route.fulfill({ status: 204, headers: corsHeaders });
      return;
    }
    fake.requests.push(
      `${request.method()} ${new URL(request.url()).pathname}`,
    );
    if (request.url().endsWith("/environment")) {
      await (fake.environment?.(route) ??
        answerJson(route, surveyEnvironment(survey)));
      return;
    }
    await answerJson(route, { data: { id: "response-1" } });
  });

  return fake;
}

async function outcomeOf(username: string) {
  const { token } = await loginAs(username, password);
  const { data } = await getUserInfo(buildRequestContext({ token }));
  return data.onboarding_survey;
}

const surveyHeading = (page: Page) =>
  page.getByRole("heading", { name: "Tell us about you" });

const firstRunStep = (page: Page) =>
  page.getByRole("heading", { name: "Install the agent" });

test.describe("the onboarding survey", () => {
  test.skip(isEnterprise, enterpriseReason);

  test("a new user answers it on their first sign-in and is not asked again", async ({
    page,
  }) => {
    const team = await createTeam();
    const fake = await serveFakeSurvey(page);

    await signIn(page, team.owner.username, password);

    await expect(surveyHeading(page)).toBeVisible();
    await page.getByRole("button", { name: "Next" }).click();
    await page.getByRole("radio", { name: "Developer" }).press("Space");
    await expect(page.getByRole("radio", { name: "Developer" })).toBeChecked();
    await page.getByRole("button", { name: "Continue" }).click();
    await expect(firstRunStep(page)).toBeVisible();
    await expect
      .poll(() => fake.requests)
      .toContain("POST /api/v1/client/workspace/responses");

    await signOut(page, team.owner.username);
    await signIn(page, team.owner.username, password);
    await expect(surveyHeading(page)).toBeHidden();
    await expect(firstRunStep(page)).toBeVisible();

    await expect.poll(() => outcomeOf(team.owner.username)).toBe("answered");
  });

  test("a new user skips it on their first sign-in and is not asked again", async ({
    page,
  }) => {
    const team = await createTeam();
    await serveFakeSurvey(page);

    await signIn(page, team.owner.username, password);

    await page.getByRole("button", { name: "Skip survey" }).click();
    await expect(firstRunStep(page)).toBeVisible();

    await signOut(page, team.owner.username);
    await signIn(page, team.owner.username, password);
    await expect(surveyHeading(page)).toBeHidden();
    await expect(firstRunStep(page)).toBeVisible();

    await expect.poll(() => outcomeOf(team.owner.username)).toBe("skipped");
  });

  test("a new user reaches the console when the survey cannot be loaded, and is asked on the next sign-in", async ({
    page,
  }) => {
    const team = await createTeam();
    const fake = await serveFakeSurvey(page, {
      environment: (route) => answerJson(route, {}, 503),
    });

    await signIn(page, team.owner.username, password);
    await expect(firstRunStep(page)).toBeVisible();
    await expect(surveyHeading(page)).toBeHidden();

    fake.environment = undefined;
    await signOut(page, team.owner.username);
    await signIn(page, team.owner.username, password);

    await expect(surveyHeading(page)).toBeVisible();
  });
});

test.describe("the onboarding survey on enterprise", () => {
  test.skip(!isEnterprise, "only enterprise leaves the survey out");

  test("a new user reaches the console without being asked", async ({
    page,
  }) => {
    const team = await createTeam();
    const fake = await serveFakeSurvey(page);

    await signIn(page, team.owner.username, password);
    await expect(surveyHeading(page)).toBeHidden();
    await expect(firstRunStep(page)).toBeVisible();

    await signOut(page, team.owner.username);
    await signIn(page, team.owner.username, password);
    await expect(firstRunStep(page)).toBeVisible();

    expect(fake.requests).toEqual([]);
    await expect.poll(() => outcomeOf(team.owner.username)).toBe("pending");
  });
});
