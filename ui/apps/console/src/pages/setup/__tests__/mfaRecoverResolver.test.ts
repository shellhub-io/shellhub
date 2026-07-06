import { describe, it, expect } from "vitest";
import { mfaRecoverResolver } from "../mfaRecoverResolver";
import type { MfaRecoverFormValues } from "../mfaRecoverResolver";

const emptyOptions = {
  criteriaMode: undefined,
  fields: {},
  names: undefined,
  shouldUseNativeValidation: undefined,
} as Parameters<typeof mfaRecoverResolver>[2];

type RhfErrors = Record<keyof MfaRecoverFormValues, { type: string; message: string }>;

const resolve = (input: MfaRecoverFormValues) =>
  mfaRecoverResolver(input, undefined, emptyOptions);

async function expectFieldError(
  input: MfaRecoverFormValues,
  field: keyof MfaRecoverFormValues,
  type: string,
) {
  const result = await resolve(input);
  const errors = result.errors as Partial<RhfErrors>;

  expect(errors).toHaveProperty(field);
  expect(errors[field]?.type).toBe(type);
  expect(typeof errors[field]?.message).toBe("string");
  expect(result.values).toEqual({});
}

describe("mfaRecoverResolver", () => {
  it("accepts a non-empty recovery code", async () => {
    const result = await resolve({ recoveryCode: "ABCD-1234" });
    expect(result.errors).toEqual({});
  });

  it("rejects an empty recovery code", async () => {
    await expectFieldError({ recoveryCode: "" }, "recoveryCode", "required");
  });

  it("rejects a whitespace-only recovery code", async () => {
    await expectFieldError({ recoveryCode: "   " }, "recoveryCode", "required");
  });

  it("returns trimmed recoveryCode in values on success", async () => {
    const result = await resolve({ recoveryCode: "  ABCD-1234  " });
    expect(result.values).toEqual({ recoveryCode: "ABCD-1234" });
    expect(result.errors).toEqual({});
  });
});
