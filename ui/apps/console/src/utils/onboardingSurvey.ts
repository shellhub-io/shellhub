import { z } from "zod";

/**
 * The Formbricks action that marks a survey as the one setup shows. The survey to show is picked
 * by this trigger in the Formbricks dashboard, not by an id in each instance's config, so moving
 * the trigger to another survey reaches instances already installed.
 */
export const ONBOARDING_TRIGGER = "shellhub-setup-onboarding";

const CONTACT_FIELDS = [
  "firstName",
  "lastName",
  "email",
  "phone",
  "company",
] as const;

type ContactField = (typeof CONTACT_FIELDS)[number];

interface SurveyChoice {
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
 * A question reduced to what the console renders. These three kinds are the whole set:
 * findOnboardingSurvey returns null for a survey holding any other Formbricks question type.
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

const answersSchema = z.object({
  anonymous: z.boolean(),
  choices: z.record(z.string()),
  other: z.record(z.string()),
  contact: z.record(
    z
      .object({
        firstName: z.string(),
        lastName: z.string(),
        email: z.string(),
        phone: z.string(),
        company: z.string(),
      })
      .partial(),
  ),
  consent: z.record(z.boolean()),
});

/**
 * The answers as the survey form holds them: a choice by its id, the free text typed for an
 * "other" choice, contact fields by name, consent as a checkbox, and whether the user answers
 * anonymously, which withholds contact and consent whatever they hold.
 */
export type SurveyAnswers = z.infer<typeof answersSchema>;

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

function orderChoices(
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
 * What an account already tells us about its owner, to start the contact fields from. Either may
 * be null when the account does not say.
 */
export interface KnownContact {
  name: string | null;
  email: string | null;
}

function contactValues(
  fields: { name: ContactField }[],
  known: KnownContact,
): Partial<Record<ContactField, string>> {
  const name = (known.name ?? "").trim();
  const splitName = fields.some((f) => f.name === "lastName");
  const [first = "", ...rest] = splitName ? name.split(/\s+/) : [name];
  const values: Partial<Record<ContactField, string>> = {
    firstName: first,
    lastName: rest.join(" "),
    email: known.email ?? "",
  };
  return Object.fromEntries(fields.map((f) => [f.name, values[f.name] ?? ""]));
}

/**
 * The starting answers for a survey, the form's default values: blank, with the contact fields
 * filled from what the account tells us. The whole name goes in the first-name field unless the
 * survey also shows a last-name field, in which case it is split at the first space.
 */
export function emptyAnswers(
  survey: OnboardingSurvey,
  known: KnownContact = { name: null, email: null },
): SurveyAnswers {
  const answers: SurveyAnswers = {
    anonymous: false,
    choices: {},
    other: {},
    contact: {},
    consent: {},
  };

  for (const q of survey.questions) {
    switch (q.kind) {
      case "choice":
        answers.choices[q.id] = "";
        answers.other[q.id] = "";
        break;
      case "contact":
        answers.contact[q.id] = contactValues(q.fields, known);
        break;
      case "consent":
        answers.consent[q.id] = false;
        break;
      default:
        q satisfies never;
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
 * declares. Contact details are sent whether or not consent is given: consent decides whether we
 * may reach out, not whether we keep what was typed. An anonymous answer withholds both.
 *
 * A new response leaves out what was not answered or withheld. An update (`update: true`) sends
 * every question, blank where there is nothing to send, because Formbricks merges an update into
 * the stored data: a key left out would keep the earlier pass's value, so ticking "anonymous" on a
 * second pass would not take back the contact details sent on the first.
 */
export function responseData(
  survey: OnboardingSurvey,
  answers: SurveyAnswers,
  hidden: Record<string, string>,
  { update = false }: { update?: boolean } = {},
): Record<string, string | string[]> {
  const data: Record<string, string | string[]> = {};
  const blank = (id: string, value: string | string[]) => {
    if (update) data[id] = value;
  };

  for (const q of survey.questions) {
    switch (q.kind) {
      case "choice": {
        const choice = q.choices.find((c) => c.id === answers.choices[q.id]);
        if (!choice) {
          blank(q.id, "");
          break;
        }
        data[q.id] = isOtherChoice(choice)
          ? answers.other[q.id]?.trim() || choice.label
          : choice.label;
        break;
      }
      case "contact": {
        const values = answers.anonymous ? {} : (answers.contact[q.id] ?? {});
        const row = CONTACT_FIELDS.map((name) => values[name]?.trim() ?? "");
        if (row.some((v) => v !== "")) data[q.id] = row;
        else blank(q.id, row);
        break;
      }
      case "consent":
        if (!answers.anonymous && answers.consent[q.id]) {
          data[q.id] = "accepted";
        } else {
          blank(q.id, "");
        }
        break;
      default:
        q satisfies never;
    }
  }

  for (const id of survey.hiddenFieldIds) {
    if (hidden[id] !== undefined) data[id] = hidden[id];
  }

  return data;
}

const emailSchema = z.string().email();

/**
 * The form schema for a survey's answers: the answer shape, refined with what this survey
 * requires. A required choice needs a pick, a required contact field a value, a given email must
 * be well formed, and required consent must be ticked. Contact and consent are not checked when
 * the user answers anonymously, since they are not sent.
 */
export function surveyAnswersSchema(survey: OnboardingSurvey) {
  return answersSchema.superRefine((answers, ctx) => {
    const fail = (path: string[], message: string) =>
      ctx.addIssue({ code: z.ZodIssueCode.custom, path, message });

    for (const q of survey.questions) {
      switch (q.kind) {
        case "choice":
          if (q.required && !answers.choices[q.id]) {
            fail(["choices", q.id], "Pick one");
          }
          break;
        case "contact": {
          if (answers.anonymous) break;
          const values = answers.contact[q.id] ?? {};
          for (const field of q.fields) {
            const value = values[field.name]?.trim() ?? "";
            if (field.required && value === "") {
              fail(["contact", q.id, field.name], "Required");
            } else if (
              field.name === "email" &&
              value !== "" &&
              !emailSchema.safeParse(value).success
            ) {
              fail(["contact", q.id, field.name], "Enter a valid email");
            }
          }
          break;
        }
        case "consent":
          if (!answers.anonymous && q.required && !answers.consent[q.id]) {
            fail(["consent", q.id], "Required");
          }
          break;
        default:
          q satisfies never;
      }
    }
  });
}
