import { describe, it, expect } from "vitest";
import { updatePasswordResolver } from "../updatePasswordResolver";
import type { UpdatePasswordFormValues } from "../updatePasswordResolver";

const emptyOptions = {
  criteriaMode: undefined,
  fields: {},
  names: undefined,
  shouldUseNativeValidation: undefined,
} as Parameters<typeof updatePasswordResolver>[2];

type RhfErrors = Record<keyof UpdatePasswordFormValues, { type: string; message: string }>;

const resolve = (input: UpdatePasswordFormValues) =>
  updatePasswordResolver(input, undefined, emptyOptions);

async function expectFieldError(
  input: UpdatePasswordFormValues,
  field: keyof UpdatePasswordFormValues,
  type: string,
) {
  const result = await resolve(input);
  const errors = result.errors as Partial<RhfErrors>;

  expect(errors).toHaveProperty(field);
  expect(errors[field]?.type).toBe(type);
  expect(typeof errors[field]?.message).toBe("string");
  expect(result.values).toEqual({});
}

describe("updatePasswordResolver", () => {
  it("returns values with no errors for valid matching passwords", async () => {
    const input: UpdatePasswordFormValues = { password: "secret", confirmPassword: "secret" };
    const result = await resolve(input);
    expect(result.values).toEqual(input);
    expect(result.errors).toEqual({});
  });

  it("rejects a too-short password", async () => {
    await expectFieldError(
      { password: "abc", confirmPassword: "abc" },
      "password",
      "validate",
    );
  });

  it("rejects a too-long password", async () => {
    await expectFieldError(
      { password: "a".repeat(33), confirmPassword: "a".repeat(33) },
      "password",
      "validate",
    );
  });

  it("rejects mismatched confirmPassword with the correct message", async () => {
    const result = await resolve({ password: "secret", confirmPassword: "different" });
    const errors = result.errors as Partial<RhfErrors>;

    expect(errors).toHaveProperty("confirmPassword");
    expect(errors.confirmPassword?.message).toBe("Passwords do not match");
    expect(result.values).toEqual({});
  });
});
