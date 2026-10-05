import { ONBOARDING_TRIGGER } from "@/utils/onboardingSurvey";

/**
 * One field of a Formbricks contactInfo question, as the client API serves it.
 */
export function contactField(
  show: boolean,
  placeholder = "",
  required = false,
) {
  return { show, required, placeholder: { default: placeholder } };
}

/**
 * A Formbricks contactInfo question showing name (in the first-name field) and email, the shape
 * the onboarding survey uses. Override a field to show or hide it.
 */
export function contactElement(overrides: Record<string, unknown> = {}) {
  return {
    type: "contactInfo",
    id: "contact",
    headline: { default: "<p><b>Contact information</b></p>" },
    required: false,
    firstName: contactField(true, "Name"),
    lastName: contactField(false),
    email: contactField(true, "Email"),
    phone: contactField(false),
    company: contactField(false),
    ...overrides,
  };
}

/**
 * A required Formbricks single-choice question, its choices given as [id, label] pairs.
 */
export function choiceElement(
  id: string,
  headline: string,
  choices: [string, string][],
  overrides: Record<string, unknown> = {},
) {
  return {
    type: "multipleChoiceSingle",
    id,
    headline: { default: headline },
    required: true,
    choices: choices.map(([choiceId, label]) => ({
      id: choiceId,
      label: { default: label },
    })),
    ...overrides,
  };
}

function consentElement() {
  return {
    type: "consent",
    id: "consent",
    headline: { default: "Can we reach out?" },
    required: false,
    label: { default: "Yes, you can email me" },
  };
}

/**
 * A running Formbricks survey carrying the onboarding trigger, its questions in one block. Pass
 * `elements` to change the questions, or any survey field to override it.
 */
export function mockSurveyPayload({
  elements = [
    contactElement(),
    choiceElement("role", "What's your role?", [
      ["dev", "Developer"],
      ["ops", "Operator"],
      ["other", "Other"],
    ]),
    consentElement(),
  ],
  ...overrides
}: { elements?: unknown[] } & Record<string, unknown> = {}) {
  return {
    id: "survey-1",
    type: "app",
    status: "inProgress",
    triggers: [{ actionClass: { name: ONBOARDING_TRIGGER } }],
    hiddenFields: {
      enabled: true,
      fieldIds: ["instance_type", "instance_domain"],
    },
    blocks: [{ logic: [], elements }],
    ...overrides,
  };
}

/**
 * The Formbricks client environment payload listing the given surveys.
 */
export function surveyEnvironment(...surveys: unknown[]) {
  return { data: { data: { surveys } } };
}
