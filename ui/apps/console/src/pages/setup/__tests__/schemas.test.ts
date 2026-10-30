import { describe, expect, it } from "vitest";
import type { ZodType } from "zod";
import {
  forgotPasswordSchema,
  inviteSchema,
  loginSchema,
  mfaRecoverSchema,
  setupSchema,
  signUpSchema,
  updatePasswordSchema,
} from "../schemas";

const account = {
  name: "Admin User",
  username: "admin",
  password: "secret123",
  confirmPassword: "secret123",
};

const validSignUp = {
  ...account,
  email: "admin@example.com",
  acceptPrivacyPolicy: true,
};

const validSetup = { ...account, email: "admin@example.com", namespace: "dev" };

function failedFields(schema: ZodType, input: unknown): string[] {
  const result = schema.safeParse(input);
  return result.success ? [] : result.error.issues.map((issue) => issue.path.join(".")).sort();
}

describe("account schemas", () => {
  it.each([
    ["invite", inviteSchema, account],
    ["sign-up", signUpSchema, validSignUp],
    ["setup", setupSchema, validSetup],
  ])("accepts a valid %s form", (_, schema, input) => {
    expect(failedFields(schema, input)).toEqual([]);
  });

  it.each([
    ["name", { name: "   " }],
    ["name", { name: "a".repeat(65) }],
    ["username", { username: "ab" }],
    ["username", { username: "Admin" }],
    ["password", { password: "1234", confirmPassword: "1234" }],
    ["confirmPassword", { confirmPassword: "different" }],
  ])("rejects %s with %j on every account form", (field, change) => {
    expect(failedFields(inviteSchema, { ...account, ...change })).toEqual([field]);
    expect(failedFields(signUpSchema, { ...validSignUp, ...change })).toEqual([field]);
    expect(failedFields(setupSchema, { ...validSetup, ...change })).toEqual([field]);
  });

  it("submits name, username and email trimmed", () => {
    const parsed = signUpSchema.parse({
      ...validSignUp,
      name: "  Admin User ",
      username: " admin ",
      email: " admin@example.com ",
    });

    expect(parsed).toMatchObject({
      name: "Admin User",
      username: "admin",
      email: "admin@example.com",
    });
  });

  it("rejects an invalid email on sign-up and setup", () => {
    expect(failedFields(signUpSchema, { ...validSignUp, email: "admin@" })).toEqual(["email"]);
    expect(failedFields(setupSchema, { ...validSetup, email: "admin@" })).toEqual(["email"]);
  });

  it("requires the privacy policy on sign-up", () => {
    expect(failedFields(signUpSchema, { ...validSignUp, acceptPrivacyPolicy: false })).toEqual([
      "acceptPrivacyPolicy",
    ]);
  });

  it("checks the namespace on setup", () => {
    expect(failedFields(setupSchema, { ...validSetup, namespace: "-dev" })).toEqual(["namespace"]);
  });
});

describe("loginSchema", () => {
  it("trims the identifier", () => {
    expect(loginSchema.parse({ username: " admin ", password: "x" }).username).toBe("admin");
  });

  it("accepts an email as the identifier", () => {
    expect(failedFields(loginSchema, { username: "admin@example.com", password: "x" })).toEqual(
      [],
    );
  });

  it("reports the identifier and the password together", () => {
    expect(failedFields(loginSchema, { username: "", password: "" })).toEqual([
      "password",
      "username",
    ]);
  });
});

describe("forgotPasswordSchema", () => {
  it("trims the identifier", () => {
    expect(forgotPasswordSchema.parse({ account: " admin " }).account).toBe("admin");
  });

  it("rejects a blank identifier", () => {
    expect(failedFields(forgotPasswordSchema, { account: "  " })).toEqual(["account"]);
  });
});

describe("updatePasswordSchema", () => {
  it("accepts a matching password of valid length", () => {
    expect(failedFields(updatePasswordSchema, { password: "12345", confirmPassword: "12345" })).toEqual(
      [],
    );
  });

  it("rejects a short password", () => {
    expect(failedFields(updatePasswordSchema, { password: "1234", confirmPassword: "1234" })).toEqual(
      ["password"],
    );
  });

  it("rejects a confirmation that does not match", () => {
    expect(failedFields(updatePasswordSchema, { password: "12345", confirmPassword: "54321" })).toEqual(
      ["confirmPassword"],
    );
  });
});

describe("mfaRecoverSchema", () => {
  it("trims the code", () => {
    expect(mfaRecoverSchema.parse({ recoveryCode: " abc-123\n" }).recoveryCode).toBe("abc-123");
  });

  it("rejects a blank code", () => {
    expect(failedFields(mfaRecoverSchema, { recoveryCode: "   " })).toEqual(["recoveryCode"]);
  });
});
