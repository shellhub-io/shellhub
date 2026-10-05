import { describe, it, expect, beforeEach } from "vitest";
import {
  mergeAnswers,
  readSavedSurvey,
  writeSavedSurvey,
  type SavedSurvey,
} from "../savedSurvey";
import type { SurveyAnswers } from "../onboardingSurvey";

const answers: SurveyAnswers = {
  anonymous: false,
  choices: { role: "dev" },
  other: { role: "" },
  contact: { contact: { email: "ana@example.com" } },
  consent: { consent: true },
};

beforeEach(() => localStorage.clear());

describe("saved survey", () => {
  it("reads back the response this user saved", () => {
    const saved: SavedSurvey = { responseId: "response-1", answers };
    writeSavedSurvey("user-1", saved);

    expect(readSavedSurvey("user-1")).toEqual(saved);
    expect(readSavedSurvey("user-2")).toBeNull();
  });

  it("reads storage holding something else as nothing saved", () => {
    localStorage.setItem("onboardingSurvey:user-1", "{not json");
    expect(readSavedSurvey("user-1")).toBeNull();

    localStorage.setItem(
      "onboardingSurvey:user-1",
      JSON.stringify({ answers }),
    );
    expect(readSavedSurvey("user-1")).toBeNull();
  });

  it("keeps the blank of a question added since the answers were saved", () => {
    const base: SurveyAnswers = {
      anonymous: false,
      choices: { role: "", size: "" },
      other: { role: "", size: "" },
      contact: {},
      consent: {},
    };

    expect(mergeAnswers(base, answers).choices).toEqual({
      role: "dev",
      size: "",
    });
  });
});
