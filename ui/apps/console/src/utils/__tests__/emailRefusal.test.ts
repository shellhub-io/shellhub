import { describe, it, expect } from "vitest";
import { emailRefusalMessage } from "../emailRefusal";

describe("emailRefusalMessage", () => {
  it.each([
    ["disposable", "Use a permanent email address"],
    ["unreachable", "This domain cannot receive email. Check the spelling."],
  ])("names what to change for a %s email", (reason, message) => {
    expect(emailRefusalMessage(reason)).toBe(message);
  });

  it.each([["invalid"], ["duplicated"], [undefined]])(
    "leaves %s to the caller's own message",
    (reason) => {
      expect(emailRefusalMessage(reason)).toBeUndefined();
    },
  );
});
