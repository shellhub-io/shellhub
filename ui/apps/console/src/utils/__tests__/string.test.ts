import { describe, it, expect } from "vitest";
import { getInitials } from "../string";

describe("getInitials", () => {
  describe("space-separated names", () => {
    it("returns initials from a two-word name", () => {
      expect(getInitials("John Doe")).toBe("JD");
    });

    it("returns initials from a full name with more than two words", () => {
      expect(getInitials("John Michael Doe")).toBe("JM");
    });

    it("uppercases both initials", () => {
      expect(getInitials("alice bob")).toBe("AB");
    });
  });

  describe("email addresses (split on @)", () => {
    it("returns initials from the local part and domain", () => {
      expect(getInitials("john@example.com")).toBe("JE");
    });

    it("handles an email where the local part contains dots", () => {
      expect(getInitials("john.doe@example.com")).toBe("JD");
    });
  });

  describe("namespace-style names with hyphens and underscores", () => {
    it("splits on hyphens", () => {
      expect(getInitials("my-namespace")).toBe("MN");
    });

    it("splits on underscores", () => {
      expect(getInitials("my_project")).toBe("MP");
    });

    it("splits on dots", () => {
      expect(getInitials("org.team")).toBe("OT");
    });

    it("handles mixed delimiters", () => {
      expect(getInitials("my-cool_project")).toBe("MC");
    });
  });

  describe("single word", () => {
    it("returns only the first initial when there is one word", () => {
      expect(getInitials("Admin")).toBe("A");
    });

    it("uppercases a lowercase single word", () => {
      expect(getInitials("admin")).toBe("A");
    });
  });

  describe("edge cases", () => {
    it("returns an empty string for an empty input", () => {
      expect(getInitials("")).toBe("");
    });

    it("handles leading and trailing delimiters gracefully", () => {
      // split produces empty strings at the boundaries — w[0] is undefined, ?? "" handles it
      expect(getInitials("-leading")).toBe("L");
    });

    it("collapses consecutive delimiters", () => {
      // The regex [\s\-_@.]+ matches one-or-more delimiters as a single separator,
      // so "a  b" splits into ["a", "b"] and both initials are returned.
      expect(getInitials("a  b")).toBe("AB");
    });
  });
});
