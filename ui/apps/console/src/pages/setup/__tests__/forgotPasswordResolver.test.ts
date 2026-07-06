import { describe, it, expect } from "vitest";
import { forgotPasswordResolver } from "../forgotPasswordResolver";
import type { ForgotPasswordFormValues } from "../forgotPasswordResolver";

const emptyOptions = {
  criteriaMode: undefined,
  fields: {},
  names: undefined,
  shouldUseNativeValidation: undefined,
} as Parameters<typeof forgotPasswordResolver>[2];

type RhfErrors = Record<keyof ForgotPasswordFormValues, { type: string; message: string }>;

const resolve = (input: ForgotPasswordFormValues) =>
  forgotPasswordResolver(input, undefined, emptyOptions);

async function expectFieldError(
  input: ForgotPasswordFormValues,
  field: keyof ForgotPasswordFormValues,
  type: string,
) {
  const result = await resolve(input);
  const errors = result.errors as Partial<RhfErrors>;

  expect(errors).toHaveProperty(field);
  expect(errors[field]?.type).toBe(type);
  expect(typeof errors[field]?.message).toBe("string");
  expect(result.values).toEqual({});
}

describe("forgotPasswordResolver", () => {
  it("accepts a valid username", async () => {
    const result = await resolve({ account: "admin" });
    expect(result.errors).toEqual({});
  });

  it("accepts a valid email", async () => {
    const result = await resolve({ account: "user@example.com" });
    expect(result.errors).toEqual({});
  });

  it("rejects an invalid value", async () => {
    await expectFieldError({ account: "!!" }, "account", "validate");
  });

  it("returns trimmed account in values on success", async () => {
    const result = await resolve({ account: "  admin  " });
    expect(result.values).toEqual({ account: "admin" });
    expect(result.errors).toEqual({});
  });
});
