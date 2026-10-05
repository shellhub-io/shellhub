import { z } from "zod";

/**
 * The Formbricks action that marks a survey as the one setup shows. The survey to show is picked
 * by this trigger in the Formbricks dashboard, not by an id in each instance's config, so moving
 * the trigger to another survey reaches instances already installed.
 */
export const ONBOARDING_TRIGGER = "shellhub-setup-onboarding";

/**
 * The contact fields in the order Formbricks stores a contactInfo answer.
 */
export const CONTACT_FIELDS = [
  "firstName",
  "lastName",
  "email",
  "phone",
  "company",
] as const;

/**
 * One of the contact fields a contactInfo question can ask for.
 */
export type ContactField = (typeof CONTACT_FIELDS)[number];

/**
 * One option of a single-choice question, with the label shown and sent as the answer.
 */
export interface SurveyChoice {
  id: string;
  label: string;
}

interface QuestionBase {
  id: string;
  headline: string;
  subheader: string;
  required: boolean;
}

/**
 * A question setup knows how to render: contact details, a single choice, or a consent box.
 */
export type SurveyQuestion =
  | (QuestionBase & {
      kind: "contact";
      fields: { name: ContactField; placeholder: string; required: boolean }[];
    })
  | (QuestionBase & { kind: "choice"; choices: SurveyChoice[] })
  | (QuestionBase & { kind: "consent"; label: string });

/**
 * A survey reduced to what setup can render. Only built when every question is of a kind setup
 * knows and no question carries branching logic.
 */
export interface OnboardingSurvey {
  id: string;
  hiddenFieldIds: string[];
  questions: SurveyQuestion[];
}

/**
 * The answers as the setup form holds them: a choice by its id, the free text typed for an
 * "other" choice, contact fields by name, consent as a checkbox.
 */
export interface SurveyAnswers {
  choices: Record<string, string>;
  other: Record<string, string>;
  contact: Record<string, Partial<Record<ContactField, string>>>;
  consent: Record<string, boolean>;
}

const OTHER_CHOICE_ID = "other";

const i18nString = z.object({ default: z.string() });

const contactFieldSchema = z.object({
  show: z.boolean(),
  required: z.boolean(),
  placeholder: i18nString.optional(),
});

const elementSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("contactInfo"),
    id: z.string(),
    headline: i18nString,
    subheader: i18nString.optional(),
    required: z.boolean(),
    firstName: contactFieldSchema,
    lastName: contactFieldSchema,
    email: contactFieldSchema,
    phone: contactFieldSchema,
    company: contactFieldSchema,
  }),
  z.object({
    type: z.literal("multipleChoiceSingle"),
    id: z.string(),
    headline: i18nString,
    subheader: i18nString.optional(),
    required: z.boolean(),
    shuffleOption: z.string().optional(),
    choices: z.array(z.object({ id: z.string(), label: i18nString })),
  }),
  z.object({
    type: z.literal("consent"),
    id: z.string(),
    headline: i18nString,
    subheader: i18nString.optional(),
    required: z.boolean(),
    label: i18nString,
  }),
]);

type SurveyElement = z.infer<typeof elementSchema>;

const blockSchema = z.object({
  logic: z.array(z.unknown()).nullish(),
  elements: z.array(z.unknown()),
});

const surveySchema = z.object({
  id: z.string(),
  type: z.string(),
  status: z.string(),
  triggers: z
    .array(z.object({ actionClass: z.object({ name: z.string() }) }))
    .default([]),
  hiddenFields: z
    .object({ enabled: z.boolean(), fieldIds: z.array(z.string()).optional() })
    .optional(),
  blocks: z.array(blockSchema).default([]),
});

const environmentSchema = z.object({
  data: z.object({
    data: z.object({ surveys: z.array(z.unknown()) }),
  }),
});

function plainText(html: string | undefined): string {
  if (!html) return "";
  const doc = new DOMParser().parseFromString(html, "text/html");
  return (doc.body.textContent ?? "").trim();
}

/**
 * Orders the choices the way the survey asks: "all" shuffles every choice, "exceptLast" keeps the
 * last one (usually "Other") in place. Answers drift toward whatever is listed first, and a
 * shuffle spreads that bias across the choices instead of piling it on one.
 */
export function orderChoices(
  choices: SurveyChoice[],
  shuffle: string | undefined,
  random: () => number = Math.random,
): SurveyChoice[] {
  if (shuffle !== "all" && shuffle !== "exceptLast") return choices;

  const pinned = shuffle === "exceptLast" ? choices.slice(-1) : [];
  const shuffled = choices.slice(0, choices.length - pinned.length);

  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }

  return [...shuffled, ...pinned];
}

function toQuestion(
  element: SurveyElement,
  random: () => number,
): SurveyQuestion {
  const base = {
    id: element.id,
    headline: plainText(element.headline.default),
    subheader: plainText(element.subheader?.default),
    required: element.required,
  };

  switch (element.type) {
    case "contactInfo":
      return {
        ...base,
        kind: "contact",
        fields: CONTACT_FIELDS.filter((name) => element[name].show).map(
          (name) => ({
            name,
            placeholder: element[name].placeholder?.default ?? "",
            required: element[name].required,
          }),
        ),
      };
    case "multipleChoiceSingle":
      return {
        ...base,
        kind: "choice",
        choices: orderChoices(
          element.choices.map((c) => ({ id: c.id, label: c.label.default })),
          element.shuffleOption,
          random,
        ),
      };
    case "consent":
      return { ...base, kind: "consent", label: element.label.default };
  }
}

/**
 * Picks the survey carrying the onboarding trigger out of a Formbricks client environment
 * payload. Returns null when there is none, when it is not running, or when it uses a question
 * type or branching that setup does not render, so setup skips the survey rather than showing
 * part of it.
 */
export function findOnboardingSurvey(
  payload: unknown,
  random: () => number = Math.random,
): OnboardingSurvey | null {
  const environment = environmentSchema.safeParse(payload);
  if (!environment.success) return null;

  for (const raw of environment.data.data.data.surveys) {
    const survey = surveySchema.safeParse(raw);
    if (!survey.success) continue;

    const { id, status, triggers, hiddenFields, blocks } = survey.data;
    if (!triggers.some((t) => t.actionClass.name === ONBOARDING_TRIGGER)) {
      continue;
    }
    if (status !== "inProgress") return null;
    if (blocks.some((b) => (b.logic ?? []).length > 0)) return null;

    const elements = blocks.flatMap((b) => b.elements);
    const parsed = elements.map((e) => elementSchema.safeParse(e));
    if (parsed.length === 0 || parsed.some((p) => !p.success)) return null;

    return {
      id,
      hiddenFieldIds: hiddenFields?.enabled
        ? (hiddenFields.fieldIds ?? [])
        : [],
      questions: parsed.map((p) => toQuestion(p.data as SurveyElement, random)),
    };
  }

  return null;
}

/**
 * The empty answers for a survey, the form's default values.
 */
export function emptyAnswers(survey: OnboardingSurvey): SurveyAnswers {
  const answers: SurveyAnswers = {
    choices: {},
    other: {},
    contact: {},
    consent: {},
  };

  for (const q of survey.questions) {
    if (q.kind === "choice") {
      answers.choices[q.id] = "";
      answers.other[q.id] = "";
    } else if (q.kind === "contact") {
      answers.contact[q.id] = Object.fromEntries(
        q.fields.map((f) => [f.name, ""]),
      );
    } else {
      answers.consent[q.id] = false;
    }
  }

  return answers;
}

/**
 * Whether a choice is the free-text "other" option, which takes what the user typed as its answer.
 */
export function isOtherChoice(choice: SurveyChoice): boolean {
  return choice.id === OTHER_CHOICE_ID;
}

/**
 * The response data Formbricks stores: a choice by its label (or the typed text for "other"),
 * contact info as the five-field array, consent as "accepted", and the hidden fields the survey
 * declares. Unanswered optional questions are left out.
 */
export function responseData(
  survey: OnboardingSurvey,
  answers: SurveyAnswers,
  hidden: Record<string, string>,
): Record<string, string | string[]> {
  const data: Record<string, string | string[]> = {};

  for (const q of survey.questions) {
    if (q.kind === "choice") {
      const choice = q.choices.find((c) => c.id === answers.choices[q.id]);
      if (!choice) continue;
      data[q.id] = isOtherChoice(choice)
        ? answers.other[q.id]?.trim() || choice.label
        : choice.label;
    } else if (q.kind === "contact") {
      const values = answers.contact[q.id] ?? {};
      const row = CONTACT_FIELDS.map((name) => values[name]?.trim() ?? "");
      if (row.some((v) => v !== "")) data[q.id] = row;
    } else if (answers.consent[q.id]) {
      data[q.id] = "accepted";
    }
  }

  for (const id of survey.hiddenFieldIds) {
    if (hidden[id] !== undefined) data[id] = hidden[id];
  }

  return data;
}

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * Checks the answers against what the survey requires: a pick for every required choice, the
 * required contact fields, a well-formed email when one is given, a ticked box for required
 * consent. Returns error messages keyed by form path.
 */
export function validateAnswers(
  survey: OnboardingSurvey,
  answers: SurveyAnswers,
): Record<string, string> {
  const errors: Record<string, string> = {};

  for (const q of survey.questions) {
    if (q.kind === "choice") {
      if (q.required && !answers.choices[q.id]) {
        errors[`choices.${q.id}`] = "Pick one";
      }
    } else if (q.kind === "contact") {
      const values = answers.contact[q.id] ?? {};
      for (const field of q.fields) {
        const value = values[field.name]?.trim() ?? "";
        if (field.required && value === "") {
          errors[`contact.${q.id}.${field.name}`] = "Required";
        } else if (
          field.name === "email" &&
          value !== "" &&
          !EMAIL_PATTERN.test(value)
        ) {
          errors[`contact.${q.id}.${field.name}`] = "Enter a valid email";
        }
      }
    } else if (q.required && !answers.consent[q.id]) {
      errors[`consent.${q.id}`] = "Required";
    }
  }

  return errors;
}
