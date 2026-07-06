import { describe, it, expect } from "vitest";
import { loginResolver } from "../loginResolver";
import type { LoginFormValues } from "../loginResolver";

const validInput: LoginFormValues = {
  username: "admin",
  password: "secret",
};

const emptyOptions = {
  criteriaMode: undefined,
  fields: {},
  names: undefined,
  shouldUseNativeValidation: undefined,
} as Parameters<typeof loginResolver>[2];

type RhfErrors = Record<keyof LoginFormValues, { type: string; message: string }>;

const resolve = (input: LoginFormValues) =>
  loginResolver(input, undefined, emptyOptions);

async function expectFieldError(
  input: LoginFormValues,
  field: keyof LoginFormValues,
  type: string,
) {
  const result = await resolve(input);
  const errors = result.errors as Partial<RhfErrors>;

  expect(errors).toHaveProperty(field);
  expect(errors[field]?.type).toBe(type);
  expect(typeof errors[field]?.message).toBe("string");
  expect(result.values).toEqual({});
}

describe("loginResolver", () => {
  it("accepts a valid username", async () => {
    const result = await resolve(validInput);
    expect(result.errors).toEqual({});
  });

  it("accepts a valid email as username", async () => {
    const result = await resolve({ ...validInput, username: "user@example.com" });
    expect(result.errors).toEqual({});
  });

  it("rejects an invalid username", async () => {
    await expectFieldError({ ...validInput, username: "!!" }, "username", "validate");
  });

  it("rejects an empty password", async () => {
    await expectFieldError({ ...validInput, password: "" }, "password", "required");
  });

  it("does not validate password length", async () => {
    const result = await resolve({ ...validInput, password: "a" });
    expect(result.errors).not.toHaveProperty("password");
  });

  it("returns trimmed username in values on success", async () => {
    const result = await resolve({ username: "  admin  ", password: "secret" });
    expect(result.values).toEqual({ username: "admin", password: "secret" });
    expect(result.errors).toEqual({});
  });
});
