import { describe, it, expect } from "vitest";
import {
  ONBOARDING_TRIGGER,
  emptyAnswers,
  findOnboardingSurvey,
  orderChoices,
  responseData,
  validateAnswers,
  type OnboardingSurvey,
} from "../onboardingSurvey";

const hidden = (show: boolean, required = false) => ({
  show,
  required,
  placeholder: { default: "x" },
});

function surveyPayload(overrides: Record<string, unknown> = {}) {
  return {
    id: "s1",
    type: "app",
    status: "inProgress",
    triggers: [{ actionClass: { name: ONBOARDING_TRIGGER } }],
    hiddenFields: { enabled: true, fieldIds: ["instance_domain"] },
    blocks: [
      {
        logic: [],
        elements: [
          {
            type: "contactInfo",
            id: "contact",
            headline: { default: "<p><b>Contact</b></p>" },
            required: false,
            firstName: {
              ...hidden(true),
              placeholder: { default: "First Name" },
            },
            lastName: hidden(false),
            email: { ...hidden(true), placeholder: { default: "Email" } },
            phone: hidden(false, true),
            company: hidden(false),
          },
          {
            type: "multipleChoiceSingle",
            id: "role",
            headline: { default: "Role?" },
            required: true,
            choices: [
              { id: "dev", label: { default: "Developer" } },
              { id: "other", label: { default: "Other" } },
            ],
          },
          {
            type: "consent",
            id: "consent",
            headline: { default: "Consent" },
            required: false,
            label: { default: "I agree" },
          },
        ],
      },
    ],
    ...overrides,
  };
}

function environment(...surveys: unknown[]) {
  return { data: { data: { surveys } } };
}

function parsed(): OnboardingSurvey {
  const survey = findOnboardingSurvey(environment(surveyPayload()));
  if (!survey) throw new Error("fixture survey did not parse");
  return survey;
}

describe("findOnboardingSurvey", () => {
  it("picks the survey carrying the onboarding trigger", () => {
    const other = surveyPayload({
      id: "unrelated",
      triggers: [{ actionClass: { name: "New Session" } }],
    });

    expect(findOnboardingSurvey(environment(other, surveyPayload()))?.id).toBe(
      "s1",
    );
  });

  it("reduces the questions to what setup renders", () => {
    expect(parsed().questions).toEqual([
      {
        kind: "contact",
        id: "contact",
        headline: "Contact",
        subheader: "",
        required: false,
        fields: [
          { name: "firstName", placeholder: "First Name", required: false },
          { name: "email", placeholder: "Email", required: false },
        ],
      },
      {
        kind: "choice",
        id: "role",
        headline: "Role?",
        subheader: "",
        required: true,
        choices: [
          { id: "dev", label: "Developer" },
          { id: "other", label: "Other" },
        ],
      },
      {
        kind: "consent",
        id: "consent",
        headline: "Consent",
        subheader: "",
        required: false,
        label: "I agree",
      },
    ]);
  });

  it.each([
    ["no survey has the trigger", environment(surveyPayload({ triggers: [] }))],
    ["the survey is paused", environment(surveyPayload({ status: "paused" }))],
    [
      "a question type is unknown",
      environment(
        surveyPayload({
          blocks: [
            {
              elements: [
                {
                  type: "rating",
                  id: "r",
                  headline: { default: "?" },
                  required: true,
                },
              ],
            },
          ],
        }),
      ),
    ],
    [
      "a block branches",
      environment(
        surveyPayload({
          blocks: [{ logic: [{ id: "l" }], elements: [] }],
        }),
      ),
    ],
    ["the payload is not an environment", { error: "nope" }],
  ])("returns null when %s", (_case, payload) => {
    expect(findOnboardingSurvey(payload)).toBeNull();
  });
});

describe("orderChoices", () => {
  const choices = ["a", "b", "c", "d"].map((id) => ({ id, label: id }));
  const alwaysFirst = () => 0;

  it("keeps the order when the survey does not shuffle", () => {
    expect(orderChoices(choices, "none", alwaysFirst)).toEqual(choices);
  });

  it("shuffles every choice for all", () => {
    expect(orderChoices(choices, "all", alwaysFirst).map((c) => c.id)).toEqual([
      "b",
      "c",
      "d",
      "a",
    ]);
  });

  it("keeps the order for a shuffle mode setup does not know", () => {
    expect(
      orderChoices(choices, "reverseOrderOccasionally", alwaysFirst),
    ).toEqual(choices);
  });

  it("keeps the last choice in place for exceptLast", () => {
    expect(
      orderChoices(choices, "exceptLast", alwaysFirst).map((c) => c.id),
    ).toEqual(["b", "c", "a", "d"]);
  });
});

describe("responseData", () => {
  it("sends labels, the contact row, consent and the declared hidden fields", () => {
    const survey = parsed();
    const answers = emptyAnswers(survey);
    answers.choices.role = "dev";
    answers.contact.contact = { firstName: " Ana ", email: "ana@example.com" };
    answers.consent.consent = true;

    expect(
      responseData(survey, answers, {
        instance_domain: "shellhub.example.com",
        not_declared: "x",
      }),
    ).toEqual({
      role: "Developer",
      contact: ["Ana", "", "ana@example.com", "", ""],
      consent: "accepted",
      instance_domain: "shellhub.example.com",
    });
  });

  it("sends the typed text for other, and leaves unanswered questions out", () => {
    const survey = parsed();
    const answers = emptyAnswers(survey);
    answers.choices.role = "other";
    answers.other.role = "Hobbyist";

    expect(responseData(survey, answers, {})).toEqual({ role: "Hobbyist" });
  });
});

describe("validateAnswers", () => {
  it("requires the required choice and a well-formed email", () => {
    const survey = parsed();
    const answers = emptyAnswers(survey);
    answers.contact.contact = { email: "not-an-email" };

    expect(validateAnswers(survey, answers)).toEqual({
      "choices.role": "Pick one",
      "contact.contact.email": "Enter a valid email",
    });
  });

  it("passes once the required answers are in", () => {
    const survey = parsed();
    const answers = emptyAnswers(survey);
    answers.choices.role = "dev";

    expect(validateAnswers(survey, answers)).toEqual({});
  });
});
