import { describe, it, expect } from "vitest";
import { setupResolver } from "../setupResolver";
import type { SetupFormValues } from "../setupResolver";

const validInput: SetupFormValues = {
  name: "Admin User",
  username: "admin",
  email: "admin@example.com",
  password: "secret123",
  confirmPassword: "secret123",
};

const emptyOptions = {
  criteriaMode: undefined,
  fields: {},
  names: undefined,
  shouldUseNativeValidation: undefined,
} as Parameters<typeof setupResolver>[2];

type RhfErrors = Record<keyof SetupFormValues, { type: string; message: string }>;

const resolve = (input: SetupFormValues) =>
  setupResolver(input, undefined, emptyOptions);

/**
 * Every invalid-field case shares the same shape: the resolver should surface a
 * single RHF error on `field` with the expected `type`, a string message, and
 * empty `values`. Asserting that contract in one place keeps the cases below to
 * just the input that triggers each error.
 */
async function expectFieldError(
  input: SetupFormValues,
  field: keyof SetupFormValues,
  type: string,
) {
  const result = await resolve(input);
  const errors = result.errors as Partial<RhfErrors>;

  expect(errors).toHaveProperty(field);
  expect(errors[field]?.type).toBe(type);
  expect(typeof errors[field]?.message).toBe("string");
  expect(result.values).toEqual({});
}

describe("setupResolver", () => {
  it("returns values with no errors for valid input", async () => {
    const result = await resolve(validInput);
    expect(result.values).toEqual(validInput);
    expect(result.errors).toEqual({});
  });

  describe("validate() field error mapping", () => {
    const invalidCases: { field: keyof SetupFormValues; input: SetupFormValues }[] = [
      { field: "name", input: { ...validInput, name: "" } },
      { field: "username", input: { ...validInput, username: "ab" } },
      { field: "email", input: { ...validInput, email: "not-an-email" } },
      { field: "password", input: { ...validInput, password: "123", confirmPassword: "123" } },
      { field: "confirmPassword", input: { ...validInput, confirmPassword: "different" } },
    ];

    it.each(invalidCases)("maps $field error to RHF error shape", async ({ field, input }) => {
      await expectFieldError(input, field, "validate");
    });
  });
});
