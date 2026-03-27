import { describe, it, expect } from "vitest";
import { isSdkError } from "../errors";

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
