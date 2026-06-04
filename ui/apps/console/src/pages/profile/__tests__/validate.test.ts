import { describe, it, expect } from "vitest";
import { validateRecoveryEmail } from "../validate";

describe("validateRecoveryEmail", () => {
  it("returns null when recovery email is empty", () => {
    expect(validateRecoveryEmail("", "user@example.com")).toBeNull();
  });

  it("rejects invalid email format", () => {
    expect(validateRecoveryEmail("not-an-email", "user@example.com")).toBe("Invalid email format");
  });

  it("rejects when emails match exactly", () => {
    expect(validateRecoveryEmail("user@example.com", "user@example.com")).toBe("Must be different from your email");
  });

  it("rejects when emails match case-insensitively", () => {
    expect(validateRecoveryEmail("User@Example.COM", "user@example.com")).toBe("Must be different from your email");
  });

  it("rejects when primary email has mixed case", () => {
    expect(validateRecoveryEmail("user@example.com", "User@Example.COM")).toBe("Must be different from your email");
  });

  it("returns null for a valid different email", () => {
    expect(validateRecoveryEmail("other@example.com", "user@example.com")).toBeNull();
  });
});
