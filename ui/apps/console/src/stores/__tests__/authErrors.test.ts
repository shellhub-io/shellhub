import { describe, it, expect } from "vitest";
import {
  MfaTokenMissingError,
  mfaCodeErrorMessage,
  mfaResetRequestError,
  recoveryCodeErrorMessage,
} from "../authErrors";

describe("authErrors", () => {
  describe("mfaCodeErrorMessage", () => {
    it.each([
      [{ status: 429 }, "Too many failed attempts. Please wait before trying again."],
      [{ status: 401 }, "Invalid verification code"],
      [new Error("network"), "Invalid verification code"],
    ])("maps %o to %s", (err, message) => {
      expect(mfaCodeErrorMessage(err)).toBe(message);
    });
  });

  describe("recoveryCodeErrorMessage", () => {
    it.each([
      [new MfaTokenMissingError(), "Sign in with your password first."],
      [{ status: 403 }, "Invalid recovery code"],
      [new Error("network"), "Invalid recovery code"],
    ])("maps %o to %s", (err, message) => {
      expect(recoveryCodeErrorMessage(err)).toBe(message);
    });
  });

  describe("mfaResetRequestError", () => {
    it.each([
      [
        { status: 429 },
        "A reset was already requested in the last minute. Check both inboxes, or wait a minute and try again.",
      ],
      [{ status: 500 }, "Unable to send reset emails. Please try again."],
      [new Error("network"), "Unable to send reset emails. Please try again."],
    ])("maps %o to %s", (err, message) => {
      expect(mfaResetRequestError(err)).toBe(message);
    });
  });
});
