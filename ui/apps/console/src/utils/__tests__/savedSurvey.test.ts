import { describe, it, expect, beforeEach } from "vitest";
import { readSavedResponseId, saveResponseId } from "../savedSurvey";

beforeEach(() => localStorage.clear());

describe("saved survey response", () => {
  it("reads back the response id only for the user and survey it was saved for", () => {
    saveResponseId("user-1", "survey-1", "response-1");

    expect(readSavedResponseId("user-1", "survey-1")).toBe("response-1");
    expect(readSavedResponseId("user-2", "survey-1")).toBeNull();
    expect(readSavedResponseId("user-1", "survey-2")).toBeNull();
  });

  it("keeps nothing for a user without an id", () => {
    saveResponseId(null, "survey-1", "response-1");

    expect(readSavedResponseId(null, "survey-1")).toBeNull();
    expect(localStorage.length).toBe(0);
  });
});
