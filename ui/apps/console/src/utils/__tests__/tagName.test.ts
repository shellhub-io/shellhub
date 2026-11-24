import { describe, expect, it } from "vitest";
import { tagNameProblem } from "@/utils/tagName";

describe("tagNameProblem", () => {
  it.each([
    { what: "3 characters", name: "abc" },
    { what: "255 characters", name: "a".repeat(255) },
    { what: "mixed letters and numbers", name: "Web01" },
  ])("accepts a name of $what", ({ name }) => {
    expect(tagNameProblem(name)).toBeNull();
  });

  it.each([
    { what: "the empty name", name: "", hint: "At least 3 characters", incomplete: true },
    { what: "a 2-character name", name: "ab", hint: "At least 3 characters", incomplete: true },
    { what: "a short name with a dash", name: "a-", hint: "At least 3 characters", incomplete: false },
    { what: "a 256-character name", name: "a".repeat(256), hint: "At most 255 characters", incomplete: false },
    { what: "a name with a dash", name: "web-01", hint: "Only letters and numbers", incomplete: false },
    { what: "a 256-character name with a dash", name: `${"a".repeat(255)}-`, hint: "At most 255 characters", incomplete: false },
  ])('answers "$hint" for $what', ({ name, hint, incomplete }) => {
    expect(tagNameProblem(name)).toEqual({ hint, incomplete });
  });
});
