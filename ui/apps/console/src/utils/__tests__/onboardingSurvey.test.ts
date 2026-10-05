import { describe, it, expect } from "vitest";
import {
  choiceElement,
  contactElement,
  contactField,
  mockSurveyPayload,
  surveyEnvironment,
} from "@/tests/onboardingSurvey";
import {
  emptyAnswers,
  findOnboardingSurvey,
  responseData,
  surveyAnswersSchema,
  type OnboardingSurvey,
  type SurveyAnswers,
} from "../onboardingSurvey";

function parse(payload = mockSurveyPayload()): OnboardingSurvey {
  const survey = findOnboardingSurvey(surveyEnvironment(payload));
  if (!survey) throw new Error("fixture survey did not parse");
  return survey;
}

function errorsOf(
  survey: OnboardingSurvey,
  answers: SurveyAnswers,
): Record<string, string> {
  const result = surveyAnswersSchema(survey).safeParse(answers);
  if (result.success) return {};
  return Object.fromEntries(
    result.error.issues.map((i) => [i.path.join("."), i.message]),
  );
}

function choiceIds(survey: OnboardingSurvey): string[] {
  const question = survey.questions.find((q) => q.kind === "choice");
  return question?.kind === "choice" ? question.choices.map((c) => c.id) : [];
}

describe("findOnboardingSurvey", () => {
  it("picks the survey carrying the onboarding trigger", () => {
    const unrelated = mockSurveyPayload({
      id: "unrelated",
      triggers: [{ actionClass: { name: "New Session" } }],
    });

    expect(
      findOnboardingSurvey(surveyEnvironment(unrelated, mockSurveyPayload()))
        ?.id,
    ).toBe("survey-1");
  });

  it("reduces the questions to what the console renders", () => {
    expect(parse().questions).toEqual([
      {
        kind: "contact",
        id: "contact",
        headline: "Contact information",
        subheader: "",
        required: false,
        fields: [
          { name: "firstName", placeholder: "Name", required: false },
          { name: "email", placeholder: "Email", required: false },
        ],
      },
      {
        kind: "choice",
        id: "role",
        headline: "What's your role?",
        subheader: "",
        required: true,
        choices: [
          { id: "dev", label: "Developer" },
          { id: "ops", label: "Operator" },
          { id: "other", label: "Other" },
        ],
      },
      {
        kind: "consent",
        id: "consent",
        headline: "Can we reach out?",
        subheader: "",
        required: false,
        label: "Yes, you can email me",
      },
    ]);
  });

  it.each([
    [
      "no survey has the trigger",
      surveyEnvironment(mockSurveyPayload({ triggers: [] })),
    ],
    [
      "the survey is paused",
      surveyEnvironment(mockSurveyPayload({ status: "paused" })),
    ],
    [
      "a question type is unknown",
      surveyEnvironment(
        mockSurveyPayload({
          elements: [
            {
              type: "rating",
              id: "r",
              headline: { default: "?" },
              required: true,
            },
          ],
        }),
      ),
    ],
    [
      "a block branches",
      surveyEnvironment(
        mockSurveyPayload({ blocks: [{ logic: [{ id: "l" }], elements: [] }] }),
      ),
    ],
    ["the payload is not an environment", { error: "nope" }],
  ])("returns null when %s", (_case, payload) => {
    expect(findOnboardingSurvey(payload)).toBeNull();
  });

  describe("choice order", () => {
    const fourChoices = (shuffleOption: string) =>
      mockSurveyPayload({
        elements: [
          choiceElement(
            "q",
            "?",
            [
              ["a", "A"],
              ["b", "B"],
              ["c", "C"],
              ["d", "D"],
            ],
            { shuffleOption },
          ),
        ],
      });
    const alwaysFirst = () => 0;
    const order = (shuffleOption: string) => {
      const survey = findOnboardingSurvey(
        surveyEnvironment(fourChoices(shuffleOption)),
        alwaysFirst,
      );
      if (!survey) throw new Error("fixture survey did not parse");
      return choiceIds(survey);
    };

    it.each([
      ["none", ["a", "b", "c", "d"]],
      ["all", ["b", "c", "d", "a"]],
      ["exceptLast", ["b", "c", "a", "d"]],
      ["reverseOrderOccasionally", ["a", "b", "c", "d"]],
    ])("orders the choices for shuffle option %s", (option, expected) => {
      expect(order(option)).toEqual(expected);
    });
  });
});

describe("emptyAnswers", () => {
  it("puts the whole account name in the name field when the survey hides last name", () => {
    expect(
      emptyAnswers(parse(), {
        name: "Ana Maria Souza",
        email: "ana@example.com",
      }).contact.contact,
    ).toEqual({ firstName: "Ana Maria Souza", email: "ana@example.com" });
  });

  it("splits the account name at the first space when the survey shows last name", () => {
    const survey = parse(
      mockSurveyPayload({
        elements: [
          contactElement({ lastName: contactField(true, "Last name") }),
        ],
      }),
    );

    expect(
      emptyAnswers(survey, { name: "Ana Maria Souza", email: null }).contact
        .contact,
    ).toEqual({ firstName: "Ana", lastName: "Maria Souza", email: "" });
  });
});

describe("responseData", () => {
  it("sends labels, the contact row, consent and the declared hidden fields", () => {
    const survey = parse();
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

  it("sends the contact details even without consent", () => {
    const survey = parse();
    const answers = emptyAnswers(survey);
    answers.choices.role = "dev";
    answers.contact.contact = { email: "ana@example.com" };

    expect(responseData(survey, answers, {})).toEqual({
      role: "Developer",
      contact: ["", "", "ana@example.com", "", ""],
    });
  });

  it("leaves out contact and consent when the user answers anonymously", () => {
    const survey = parse();
    const answers = emptyAnswers(survey);
    answers.anonymous = true;
    answers.choices.role = "dev";
    answers.contact.contact = { email: "ana@example.com" };
    answers.consent.consent = true;

    expect(responseData(survey, answers, {})).toEqual({ role: "Developer" });
  });

  it("blanks every unanswered or withheld question on an update", () => {
    const survey = parse();
    const answers = emptyAnswers(survey);
    answers.anonymous = true;
    answers.choices.role = "dev";
    answers.contact.contact = { email: "ana@example.com" };
    answers.consent.consent = true;

    expect(responseData(survey, answers, {}, { update: true })).toEqual({
      role: "Developer",
      contact: ["", "", "", "", ""],
      consent: "",
    });
  });

  it("sends the typed text for other, and leaves unanswered questions out", () => {
    const survey = parse();
    const answers = emptyAnswers(survey);
    answers.choices.role = "other";
    answers.other.role = "Hobbyist";

    expect(responseData(survey, answers, {})).toEqual({ role: "Hobbyist" });
  });
});

describe("surveyAnswersSchema", () => {
  it("requires the required choice and a well-formed email", () => {
    const survey = parse();
    const answers = emptyAnswers(survey);
    answers.contact.contact = { email: "not-an-email" };

    expect(errorsOf(survey, answers)).toEqual({
      "choices.role": "Pick one",
      "contact.contact.email": "Enter a valid email",
    });
  });

  it("does not check contact fields the user withholds by answering anonymously", () => {
    const survey = parse();
    const answers = emptyAnswers(survey);
    answers.anonymous = true;
    answers.choices.role = "dev";
    answers.contact.contact = { email: "not-an-email" };

    expect(errorsOf(survey, answers)).toEqual({});
  });

  it("passes once the required answers are in", () => {
    const survey = parse();
    const answers = emptyAnswers(survey);
    answers.choices.role = "dev";

    expect(errorsOf(survey, answers)).toEqual({});
  });
});
