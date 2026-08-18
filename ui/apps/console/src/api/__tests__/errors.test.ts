import { describe, it, expect } from "vitest";
import { apiErrorFields, apiErrorMessage, isSdkError } from "../errors";

describe("isSdkError", () => {
  describe("returns true for valid SDK errors", () => {
    it("returns true when object has numeric status property", () => {
      expect(isSdkError({ status: 400 })).toBe(true);
    });

    it("returns true for status 200", () => {
      expect(isSdkError({ status: 200 })).toBe(true);
    });

    it("returns true for status 500", () => {
      expect(isSdkError({ status: 500, headers: new Headers() })).toBe(true);
    });

    it("returns true when extra properties are present", () => {
      expect(isSdkError({ status: 401, extra: true })).toBe(true);
    });

    it("returns true for enriched arrays (real SDK shape)", () => {
      expect(
        isSdkError(Object.assign(["username"], { status: 400, headers: new Headers() })),
      ).toBe(true);
    });
  });

  describe("returns false for non-SDK errors", () => {
    it("returns false for null", () => {
      expect(isSdkError(null)).toBe(false);
    });

    it("returns false for undefined", () => {
      expect(isSdkError(undefined)).toBe(false);
    });

    it("returns false for a plain string", () => {
      expect(isSdkError("error")).toBe(false);
    });

    it("returns false for a number", () => {
      expect(isSdkError(42)).toBe(false);
    });

    it("returns false for an object missing status", () => {
      expect(isSdkError({ code: 404 })).toBe(false);
    });

    it("returns false when status is a string instead of a number", () => {
      expect(isSdkError({ status: "400" })).toBe(false);
    });

    it("returns false for an empty object", () => {
      expect(isSdkError({})).toBe(false);
    });

    it("returns false for a plain Error instance without status", () => {
      expect(isSdkError(new Error("oops"))).toBe(false);
    });

    it("returns false for a plain array without status", () => {
      expect(isSdkError(["username"])).toBe(false);
    });
  });
});

describe("apiErrorMessage", () => {
  it.each([
    [400, "Some values are invalid. Review the form and try again."],
    [401, "Your session has expired. Sign in again."],
    [402, "This action needs an active subscription."],
    [403, "You do not have permission to do this."],
    [404, "We could not find what you asked for."],
    [409, "That value is already in use."],
    [423, "This account is not active yet."],
    [429, "Too many attempts. Wait a moment and try again."],
    [500, "Something went wrong on our side. Try again."],
    [503, "The service is unavailable right now. Try again shortly."],
  ])("maps status %i to its own copy", (status, expected) => {
    expect(apiErrorMessage({ status })).toBe(expected);
  });

  it("falls back to a generic message for an unmapped status", () => {
    expect(apiErrorMessage({ status: 418 })).toBe("Something went wrong. Please try again.");
  });

  it("falls back to a generic message for a value that is not an API error", () => {
    expect(apiErrorMessage(new Error("boom"))).toBe("Something went wrong. Please try again.");
  });

  it("never renders the server's own message", () => {
    expect(apiErrorMessage({ status: 409, message: "user duplicated" })).toBe(
      "That value is already in use.",
    );
  });
});

describe("apiErrorFields", () => {
  it("returns the field map when the body carries one", () => {
    expect(apiErrorFields({ status: 409, fields: { username: "duplicated" } })).toEqual({
      username: "duplicated",
    });
  });

  it("returns an empty map when the body carries no fields", () => {
    expect(apiErrorFields({ status: 409, message: "user duplicated" })).toEqual({});
  });

  it("returns an empty map for a value that is not an API error", () => {
    expect(apiErrorFields(new Error("boom"))).toEqual({});
  });

  it("drops members whose value is not a string", () => {
    expect(
      apiErrorFields({ status: 400, fields: { username: "required", age: 42, email: null } }),
    ).toEqual({ username: "required" });
  });

  it("returns an empty map when fields is not an object", () => {
    expect(apiErrorFields({ status: 400, fields: "username" })).toEqual({});
  });
});
